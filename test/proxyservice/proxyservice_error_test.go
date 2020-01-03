package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

const SecretlessImageName = "secretless-broker"

func dockerContainer(imageName string) (types.Container, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return types.Container{}, err
	}

	cli.NegotiateAPIVersion(ctx)

	containerListOptions := types.ContainerListOptions{
		All:    true,
		Latest: true,
	}

	containers, err := cli.ContainerList(context.Background(), containerListOptions)
	if err != nil {
		return types.Container{}, err
	}

	var brokerContainer types.Container
	for _, container := range containers {
		if container.Image != imageName {
			continue
		}
		brokerContainer = container
	}

	if brokerContainer.ID == "" {
		return types.Container{},
			fmt.Errorf("Could not find matching container for image '%s'", imageName)
	}

	return brokerContainer, nil
}

func dockerLog(container types.Container) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	cli.NegotiateAPIVersion(ctx)

	logOptions := types.ContainerLogsOptions{
		Timestamps: false,
		ShowStdout: true,
	}

	out, err := cli.ContainerLogs(ctx, container.ID, logOptions)
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
