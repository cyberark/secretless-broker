package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testIODeadline = 2 * time.Second

// readNLines reads n lines from net.Conn using testIODeadline as the read
// deadline to avoid tests that hang forever.
func readNLines(conn net.Conn, n int) ([]string, error) {
	defer func() {
		_ = conn.SetReadDeadline(time.Time{})
	}()

	var lines []string
	lineReader := bufio.NewReader(conn)

	for linesRead := 0; linesRead < n; linesRead++ {
		_ = conn.SetReadDeadline(time.Now().Add(testIODeadline))

		line, _, err := lineReader.ReadLine()
		if err != nil {
			err = fmt.Errorf(
				"failed reading line %d: %s",
				linesRead+1,
				err)
			return nil, err
		}

		lines = append(lines, string(line))
	}

	return lines, nil
}

// writeLine writes a line to net.Conn using testIODeadline as the write
// deadline to avoid tests that hang forever.
func writeLine(conn net.Conn, b []byte) error {
	defer func() {
		_ = conn.SetWriteDeadline(time.Time{})
	}()

	_ = conn.SetWriteDeadline(time.Now().Add(testIODeadline))
	_, err := conn.Write(append(b, '\n'))
	return err
}

func TestTCPPlugin(t *testing.T) {
	host := os.Getenv("SECRETLESS_HOST")
	if host == "" {
		host = "localhost"
	}
	address := net.JoinHostPort(host, "6175")

	// establish connection to secretless
	connection, err := net.Dial("tcp", address)
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		_ = connection.Close()
	}()

	t.Run("can inject information from credentials", func(t *testing.T) {
		// initial write
		err = writeLine(connection, []byte("hello"))
		if !assert.NoError(t, err) {
			return
		}

		// initial read
		lines, err := readNLines(connection, 2)
		if !assert.NoError(t, err) {
			return
		}
		if assert.Equal(t, []string{
			"credential injection: some secret credentials",
			"initial message from client: hello",
		}, lines) {
			return
		}
	})

	t.Run("proxies connection to target service", func(t *testing.T) {
		err = writeLine(connection, []byte("ping"))
		if !assert.NoError(t, err) {
			return
		}

		lines, err := readNLines(connection, 1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []string{"ping"}, lines)
	})
}
