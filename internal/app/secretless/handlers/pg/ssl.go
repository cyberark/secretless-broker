package pg

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"errors"
	"fmt"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
)

var (
	ErrSSLNotSupported           = errors.New("pq: SSL is not enabled on the server")
	ErrSSLKeyHasWorldPermissions = errors.New("pq: Private key file has group or world access. Permissions should be u=rw (0600) or less")
)
type values map[string]string

// ssl generates a function to upgrade a net.Conn based on the "sslmode" and
// related settings. The function is nil when no upgrade should take place.
func ssl(connection net.Conn, o values) (net.Conn, error) {
	verifyCaOnly := false
	tlsConf := tls.Config{}
	switch mode := o["sslmode"]; mode {
	// "disable" is the default.
	case "", "disable":
		return connection, nil
	case "require":
		// We must skip TLS's own verification since it requires full
		// verification since Go 1.3.
		tlsConf.InsecureSkipVerify = true

		// From http://www.postgresql.org/docs/current/static/libpq-ssl.html:
		//
		// Note: For backwards compatibility with earlier versions of
		// PostgreSQL, if a root CA file exists, the behavior of
		// sslmode=require will be the same as that of verify-ca, meaning the
		// server certificate is validated against the CA. Relying on this
		// behavior is discouraged, and applications that need certificate
		// validation should always use verify-ca or verify-full.
		if sslrootcert, ok := o["sslrootcert"]; ok {
			if _, err := os.Stat(sslrootcert); err == nil {
				verifyCaOnly = true
			} else {
				delete(o, "sslrootcert")
			}
		}
	case "verify-ca":
		// We must skip TLS's own verification since it requires full
		// verification since Go 1.3.
		tlsConf.InsecureSkipVerify = true
		verifyCaOnly = true
	case "verify-full":
		tlsConf.ServerName = o["host"]
	default:
		return nil, fmt.Errorf(`unsupported sslmode %q; only "require" (default), "verify-full", "verify-ca", and "disable" supported`, mode)
	}

	// Start SSL Check
	/*
	* First determine if SSL is allowed by the backend. To do this, send an
	* SSL request. The response from the backend will be a single byte
	* message. If the value is 'S', then SSL connections are allowed and an
	* upgrade to the connection should be attempted. If the value is 'N',
	* then the backend does not support SSL connections.
	*/

	/* Create the SSL request message. */
	message := protocol.NewMessageBuffer([]byte{})
	message.WriteInt32(8)
	message.WriteInt32(protocol.SSLRequestCode)

	/* Send the SSL request message. */
	_, err := connection.Write(message.Bytes())

	if err != nil {
		return nil, err
	}

	/* Receive SSL response message. */
	response := make([]byte, 4096)
	_, err = connection.Read(response)

	if err != nil {
		return nil, err
	}

	/*
	 * If SSL is not allowed by the backend then close the connection and
	 * throw an error.
	 */
	if len(response) > 0 && response[0] != 'S' {
		fmt.Println(string(response))
		connection.Close()
		return nil, fmt.Errorf("The backend does not allow SSL connections.")
	}
	// End SSL Check

	err = sslClientCertificates(&tlsConf, o)
	if err != nil {
		return nil, err
	}
	err = sslCertificateAuthority(&tlsConf, o)
	if err != nil {
		return nil, err
	}

	// Accept renegotiation requests initiated by the backend.
	//
	// Renegotiation was deprecated then removed from PostgreSQL 9.5, but
	// the default configuration of older versions has it enabled. Redshift
	// also initiates renegotiations and cannot be reconfigured.
	tlsConf.Renegotiation = tls.RenegotiateFreelyAsClient

	client := tls.Client(connection, &tlsConf)
	if verifyCaOnly {
		err := sslVerifyCertificateAuthority(client, &tlsConf)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// sslClientCertificates adds the certificate specified in the "sslcert" and
// "sslkey" settings, or if they aren't set, from the .postgresql directory
// in the user's home directory. The configured files must exist and have
// the correct permissions.
func sslClientCertificates(tlsConf *tls.Config, o values) error {
	// user.Current() might fail when cross-compiling. We have to ignore the
	// error and continue without home directory defaults, since we wouldn't
	// know from where to load them.
	user, _ := user.Current()

	// In libpq, the client certificate is only loaded if the setting is not blank.
	//
	// https://github.com/postgres/postgres/blob/REL9_6_2/src/interfaces/libpq/fe-secure-openssl.c#L1036-L1037
	sslcert := o["sslcert"]
	if len(sslcert) == 0 && user != nil {
		sslcert = filepath.Join(user.HomeDir, ".postgresql", "postgresql.crt")
	}
	// https://github.com/postgres/postgres/blob/REL9_6_2/src/interfaces/libpq/fe-secure-openssl.c#L1045
	if len(sslcert) == 0 {
		return nil
	}
	// https://github.com/postgres/postgres/blob/REL9_6_2/src/interfaces/libpq/fe-secure-openssl.c#L1050:L1054
	if _, err := os.Stat(sslcert); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// In libpq, the ssl key is only loaded if the setting is not blank.
	//
	// https://github.com/postgres/postgres/blob/REL9_6_2/src/interfaces/libpq/fe-secure-openssl.c#L1123-L1222
	sslkey := o["sslkey"]
	if len(sslkey) == 0 && user != nil {
		sslkey = filepath.Join(user.HomeDir, ".postgresql", "postgresql.key")
	}

	if len(sslkey) > 0 {
		if err := sslKeyPermissions(sslkey); err != nil {
			return err
		}
	}

	cert, err := tls.LoadX509KeyPair(sslcert, sslkey)
	if err != nil {
		return err
	}

	tlsConf.Certificates = []tls.Certificate{cert}
	return nil
}

// sslCertificateAuthority adds the RootCA specified in the "sslrootcert" setting.
func sslCertificateAuthority(tlsConf *tls.Config, o values) error {
	// In libpq, the root certificate is only loaded if the setting is not blank.
	//
	// https://github.com/postgres/postgres/blob/REL9_6_2/src/interfaces/libpq/fe-secure-openssl.c#L950-L951
	if sslrootcert := o["sslrootcert"]; len(sslrootcert) > 0 {
		tlsConf.RootCAs = x509.NewCertPool()

		cert, err := ioutil.ReadFile(sslrootcert)
		if err != nil {
			return err
		}

		if !tlsConf.RootCAs.AppendCertsFromPEM(cert) {
			return fmt.Errorf("couldn't parse pem in sslrootcert")
		}
	}

	return nil
}

// sslVerifyCertificateAuthority carries out a TLS handshake to the server and
// verifies the presented certificate against the CA, i.e. the one specified in
// sslrootcert or the system CA if sslrootcert was not specified.
func sslVerifyCertificateAuthority(client *tls.Conn, tlsConf *tls.Config) error {
	err := client.Handshake()
	if err != nil {
		return err
	}
	certs := client.ConnectionState().PeerCertificates
	opts := x509.VerifyOptions{
		DNSName:       client.ConnectionState().ServerName,
		Intermediates: x509.NewCertPool(),
		Roots:         tlsConf.RootCAs,
	}
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}
	_, err = certs[0].Verify(opts)
	return err
}
