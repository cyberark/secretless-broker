package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
)

type options map[string]string

// DbSSLMode holds information about the DB's SSL options.
type DbSSLMode struct {
	tls.Config
	UseTLS       bool
	VerifyCaOnly bool
	Options      options
}

// NewDbSSLMode configures and creates a DbSSLMode
func NewDbSSLMode(o options, requireCanVerifyCAOnly bool) (DbSSLMode, error) {
	// NOTE for the "require" case:
	//
	// From http://www.postgresql.org/docs/current/static/libpq-ssl.html:
	//
	// Note: For backwards compatibility with earlier versions of
	// PostgreSQL, if a root CA file exists, the behavior of
	// sslmode=require will be the same as that of verify-ca, meaning the
	// server certificate is validated against the CA. Relying on this
	// behavior is discouraged, and applications that need certificate
	// validation should always use verify-ca or verify-full.
	sslMode := DbSSLMode{Options: o, UseTLS: true}

	switch mode := o["sslmode"]; mode {
	case "disable":
		sslMode.UseTLS = false
		return sslMode, nil
		// "require" is the default.
	case "", "require":
		// Skip TLS's own verification: it requires full verification since Go 1.3.
		sslMode.InsecureSkipVerify = true

		// From http://www.postgresql.org/docs/current/static/libpq-ssl.html:
		//
		// Note: For backwards compatibility with earlier versions of
		// PostgreSQL, if a root CA file exists, the behavior of
		// sslmode=require will be the same as that of verify-ca, meaning the
		// server certificate is validated against the CA. Relying on this
		// behavior is discouraged, and applications that need certificate
		// validation should always use verify-ca or verify-full.

		// MySQL on the other hand notes in its docs that it ignores
		// SSL certs if supplied in REQUIRED sslmode.
		if requireCanVerifyCAOnly && len(o["sslrootcert"]) > 0 {
			sslMode.VerifyCaOnly = true
		}
	case "verify-ca":
		// Skip TLS's own verification: it requires full verification since Go 1.3.
		sslMode.InsecureSkipVerify = true
		sslMode.VerifyCaOnly = true
	//case "verify-full":
	//	sslMode.ServerName = o["host"]
	default:
		// TODO add verify-full below
		return DbSSLMode{}, fmt.Errorf(`unsupported sslmode %q; only "require" (default), "verify-ca", and "disable" supported`, mode)
	}

	return sslMode, nil
}

// HandleSSLUpgrade upgrades a net.Conn using DbSSLMode
func HandleSSLUpgrade(connection net.Conn, tlsConf DbSSLMode) (net.Conn, error) {
	err := sslClientCertificates(&tlsConf.Config, tlsConf.Options)
	if err != nil {
		return nil, err
	}
	err = sslCertificateAuthority(&tlsConf.Config, tlsConf.Options)
	if err != nil {
		return nil, err
	}

	// Accept renegotiation requests initiated by the backend.
	//
	// Renegotiation was deprecated then removed from PostgreSQL 9.5, but
	// the default configuration of older versions has it enabled. Redshift
	// also initiates renegotiations and cannot be reconfigured.
	tlsConf.Renegotiation = tls.RenegotiateFreelyAsClient

	client := tls.Client(connection, &tlsConf.Config)
	if tlsConf.VerifyCaOnly {
		err := sslVerifyCertificateAuthority(client, &tlsConf.Config)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// sslClientCertificates adds the certificate specified in the "sslcert" and
// "sslkey" settings
func sslClientCertificates(tlsConf *tls.Config, o options) error {
	// The client certificate is only loaded if the setting is not blank.
	sslcert := o["sslcert"]
	if len(sslcert) == 0 {
		return nil
	}

	sslkey := o["sslkey"]

	certPEMBlock := []byte(sslcert)
	keyPEMBlock := []byte(sslkey)

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return err
	}

	tlsConf.Certificates = []tls.Certificate{cert}
	return nil
}

// sslCertificateAuthority adds the RootCA specified in the "sslrootcert" setting.
func sslCertificateAuthority(tlsConf *tls.Config, o options) error {
	// The root certificate is only loaded if the setting is not blank.
	if sslrootcert := o["sslrootcert"]; len(sslrootcert) > 0 {
		tlsConf.RootCAs = x509.NewCertPool()

		cert := []byte(sslrootcert)

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
