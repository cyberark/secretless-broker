package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginProvider(t *testing.T) {
	host := os.Getenv("SECRETLESS_HOST")
	if host == "" {
		host = "localhost"
	}

	expectedListener := host + ":6175"

	log.Printf("Trying to use host: %s", expectedListener)

	request := "GET / HTTP/1.1\r\n" +
		"Host: localhost:6174\r\n" +
		"User-Agent: UserAgentString\r\n" +
		"Accept: */*\r\n\r\n"

	fetchPage := func(address string) []string {
		tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
		So(err, ShouldBeNil)

		connection, err := net.DialTCP("tcp", nil, tcpAddr)
		So(err, ShouldBeNil)

		defer connection.Close()

		_, err = connection.Write([]byte(request))
		So(err, ShouldBeNil)

		rawContents, err := ioutil.ReadAll(connection)
		So(err, ShouldBeNil)

		pageContents := string(rawContents)
		return strings.Split(pageContents, "\r\n")
	}

	Convey("Can create and inject variables into the requests", t, func() {
		lines := fetchPage(expectedListener)

		So(lines, ShouldContain, "Example-Provider-Secret: exampleVariableProvider")
	})
}
