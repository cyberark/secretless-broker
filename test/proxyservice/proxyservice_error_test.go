package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

const SecretlessImageName = "secretless-broker"

func dockerContainer(imageName string) (container.Summary, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return container.Summary{}, err
	}

	cli.NegotiateAPIVersion(ctx)

	containerListOptions := container.ListOptions{
		All:    true,
		Latest: true,
	}

	containers, err := cli.ContainerList(ctx, containerListOptions)
	if err != nil {
		return container.Summary{}, err
	}

	var brokerContainer container.Summary
	for _, container := range containers {
		if container.Image != imageName {
			continue
		}
		brokerContainer = container
	}

	if brokerContainer.ID == "" {
		return container.Summary{},
			fmt.Errorf("Could not find matching container for image '%s'", imageName)
	}

	return brokerContainer, nil
}

func dockerLog(ctr container.Summary) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", err
	}

	cli.NegotiateAPIVersion(ctx)

	logOptions := container.LogsOptions{
		Timestamps: false,
		ShowStdout: true,
	}

	out, err := cli.ContainerLogs(ctx, ctr.ID, logOptions)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return buf.String(), nil
}

func TestProxyserviceErrors(t *testing.T) {
	container, err := dockerContainer(SecretlessImageName)
	if !assert.NoError(t, err) {
		return
	}

	dockerLog, err := dockerLog(container)
	if !assert.NoError(t, err) {
		return
	}

	// Check exit code
	t.Run("Fatal exit enforced", func(t *testing.T) {
		assert.Equal(t, container.State, "exited")
		assert.Contains(t, container.Status, "Exited (1)")
	})

	// TODO: Improve non-HTTP tests after https://github.com/cyberark/secretless-broker/issues/1063 is
	//       fixed.
	var failureTests = []struct {
		testType    string
		errorPrefix string
		errorSuffix string
	}{
		{"HTTP proxy", "HTTP Proxy on tcp://0.0.0.0:8080",
			"'authenticateURLsMatching' key has incorrect type"},
		{"TCP", "tcp-connector", "listen tcp: address 111111: invalid port"},
		{"SSH", "ssh-connector", "listen tcp: address 222222: invalid port"},
		{"SSH Agent", "ssh-agent-connector", "listen tcp: address 333333: invalid port"},
	}

	for _, test := range failureTests {
		t.Run(test.testType+" errors", func(t *testing.T) {
			assert.Contains(t, dockerLog, "[ERROR] Fatal error in '"+test.errorPrefix+"':")
			assert.Contains(t, dockerLog, test.errorSuffix)
		})
	}
}
