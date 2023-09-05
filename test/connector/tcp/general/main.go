package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"os/exec"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj"
)

// proxyViaSecretless issues a client request using a the 'runQuery' argument to a
// Secretless proxy service configured using the 'credentials' argument.
// proxyViaSecretless uses newInProcessProxyService to creating the in-process proxy
// service. The proxy service exists only for the lifetime of this method call.
func main() {
	// Looks like root password is encrypted using SCRAM https://www.postgresql.org/docs/current/auth-password.html
	var connectorType = "pg"
	var credentials = map[string][]byte{
		"host":     []byte("localhost"),
		"port":     []byte("32785"),
		"username": []byte("test"),
		"password": []byte("test"),
		"sslmode":  []byte("disable"),
	}
	var runClient = func(
		host string,
		port string,
	) (string, error) {
		// psql postgres://postgres:password@pg_no_tls:5432/postgres -c "select count(*) from test.test;"
		cmd := exec.Command(
			"psql",
			[]string{
				fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "postgres", "password", host, port, "postgres"),
				"-c", "select count(*) from test.test;",
			}...,
		)

		// Create a buffer to capture the output
		var outBuffer bytes.Buffer

		// Set the command's stdout and stderr to the buffer and also print it to stdout
		cmd.Stdout = io.MultiWriter(os.Stdout, &outBuffer)
		cmd.Stderr = io.MultiWriter(os.Stderr, &outBuffer)

		// Run the command
		err := cmd.Run()
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return "", errors.New(outBuffer.String())
			}

			return "", err
		}

		return outBuffer.String(), nil
	}

	internalPlugins, _ := sharedobj.GetInternalPluginsFunc()
	connectorPlugin := internalPlugins.TCPPlugins()[connectorType]

	// Create in-process proxy service
	proxyService, err := newInProcessProxyService(connectorPlugin, credentials)
	if err != nil {
		panic(err)
	}

	// Start the proxyService service
	proxyService.Start()

	// Make the client request to the proxy service
	clientResChan := make(chan Response)

	go func() {
		out, err := runClient(
			proxyService.host,
			proxyService.port,
		)

		clientResChan <- Response{
			Out: out,
			Err: err,
		}
	}()

	// Block and wait for the client response
	clientRes := <-clientResChan
	// Ensure the proxy service is stopped
	proxyService.Stop()

	if clientRes.Err != nil {
		fmt.Println()
		fmt.Println()
		fmt.Println("There was an error.")
		fmt.Println("OUT:", clientRes.Out)
		fmt.Println("ERROR:", clientRes.Err)
	}
}
