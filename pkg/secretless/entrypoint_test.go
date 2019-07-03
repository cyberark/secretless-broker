package secretless

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
)

var testManager = &plugin.Manager{}

func runEntrypoint(params *CLIParams) (stdoutOutput string, stderrOutput string) {
	// Swap our stdout with a special one for capture
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReadPipe, stdoutWritePipe, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	stderrReadPipe, stderrWritePipe, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout = stdoutWritePipe
	os.Stderr = stderrWritePipe

	stdoutOutputChan := make(chan string)
	stderrOutputChan := make(chan string)

	// Copying the output in a separate goroutine otherwise printing can block
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stdoutReadPipe)
		stdoutOutputChan <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stderrReadPipe)
		stderrOutputChan <- buf.String()
	}()

	Start(params, testManager)

	// Restore stdout
	stdoutWritePipe.Close()
	stderrWritePipe.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutOutput = <-stdoutOutputChan
	stderrOutput = <-stderrOutputChan

	return
}

func TestVersionParamShowsOutput(t *testing.T) {
	stdout, stderr := runEntrypoint(&CLIParams{
		ConfigManagerSpec: "configfile",
		ShowVersion:       true,
	})

	assert.Regexp(t, fmt.Sprintf("^secretless-broker v%s\n$", FullVersionName), stdout)
	assert.Regexp(t, "", stderr)
}
