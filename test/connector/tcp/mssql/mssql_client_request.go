package mssqltest

import (
	"fmt"
	"net"
	"strings"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
)

// localListenerOnPort creates a net.Listener at 127.0.0.1 on the given port. Note that
// passing in a port of "0" will result in a random port being used.
func localListenerOnPort(port string) (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:"+port)
}

// clientRequest is the request of an MSSQL client making a connection to a database via
// the Secretless proxyService. The fields on the struct are some of the values that the client has
// control over. Credentials are not included in this struct because those will be
// injected
type clientRequest struct {
	database string
	readOnly bool
	query    string
}

// cloneCredentials creates an independent clone of a credentials map. The resulting
// clone will not be affected by any mutations to the original, and vice-versa. The clone
// is useful for passing to a proxyService service, to avoid zeroization of the original.
func cloneCredentials(original map[string][]byte) map[string][]byte {
	credsClone := make(map[string][]byte)

	for key, value := range original {
		// Clone the value
		valueClone := make([]byte, len(value))
		copy(valueClone, value)

		// Set the key, value pair on the credentials clone
		credsClone[key] = valueClone
	}

	return credsClone
}

// proxyViaSecretless issues a client request using a the 'runQuery' argument to a
// Secretless proxy service configured using the 'credentials' argument.
// proxyViaSecretless uses newInProcessProxyService to creating the in-process proxy
// service. The proxy service exists only for the lifetime of this method call.
func (clientReq clientRequest) proxyViaSecretless(
	runQuery client.RunQuery,
	credentials map[string][]byte,
) (string, string, error) {
	// Create in-process proxy service
	proxyService, err := newInProcessProxyService(credentials)
	if err != nil {
		return "", "", err
	}

	// Ensure the proxy service is stopped
	defer proxyService.Stop()
	// Start the proxyService service
	proxyService.Start()

	// Make the client request to the proxy service
	clientResChan := runQuery.ConcurrentCall(
		client.Config{
			Host:     proxyService.host,
			Port:     proxyService.port,
			Username: "dummy",
			Password: "dummy",
			Database: clientReq.database,
			ReadOnly: clientReq.readOnly,
		},
		clientReq.query,
	)

	// Block and wait for the client response
	clientRes := <-clientResChan

	return clientRes.Out, proxyService.port, clientRes.Err
}

// proxyToCreatedMock issues a client request using a the 'runQuery' argument to a Secretless
// proxy service configured using the 'credentials' argument.
//
// NOTE: proxyToCreatedMock proxies the request to a mock server that terminates the request after the handshake
// This can have unintended effects. gomssql in particular does some weird retry, when a query is prepared!
// TODO: find out this weird gomssqldb behavior.
func (clientReq clientRequest) proxyToCreatedMock(
	runQuery client.RunQuery,
	credentials map[string][]byte,
) (*mockTargetCapture, string, error) {
	// Create mock target
	mt, err := newMockTarget("0")
	if err != nil {
		return nil, "", err
	}
	defer mt.close()

	return clientReq.proxyToMock(runQuery, credentials, mt)
}

func (clientReq clientRequest) proxyToMock(
	runQuery client.RunQuery,
	credentials map[string][]byte,
	mt *mockTarget,
) (*mockTargetCapture, string, error) {
	// Gather credentials
	baseCredentials := map[string][]byte{
		"host": []byte(mt.host),
		"port": []byte(mt.port),
	}
	for key, value := range credentials {
		baseCredentials[key] = value
	}

	// Accept on mock target
	mtResChan := mt.singleAcceptAndHandle()

	// We don't expect anything useful to come back from the client request.
	// This is a fire and forget.
	_, secretlessPort, err := clientReq.proxyViaSecretless(
		runQuery,
		baseCredentials,
	)

	mtRes := <-mtResChan

	var errStrings []string
	if mtRes.err != nil {
		// We only care about err (from the client request) if there was an error in
		// handling the request on the mock server. If this is not the case then it's
		// likely that err is because the mock server closed the connection.
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
		errStrings = append(errStrings, mtRes.err.Error())
	}

	var combinedErr error
	if len(errStrings) != 0 {
		combinedErr = fmt.Errorf("%s", strings.Join(errStrings, " AND "))
	}

	return mtRes.capture, secretlessPort, combinedErr
}
