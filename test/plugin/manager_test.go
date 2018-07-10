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

func TestManager(t *testing.T) {
	host := os.Getenv("SECRETLESS_HOST")
	if host == "" {
		host = "localhost"
	}

	passingListener := host + ":6175"
	rejectingListener := host + ":6176"

	log.Printf("Trying to use host: %s", host)

	request := "GET / HTTP/1.1\r\n" +
		"Host: localhost:6174\r\n" +
		"User-Agent: UserAgentString\r\n" +
		"Accept: */*\r\n\r\n"

	fetchPage := func(address string) ([]string, error) {
		log.Printf("Resolving...")
		tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
		if err != nil {
			return nil, err
		}

		log.Printf("Connecting...")
		connection, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return nil, err
		}

		defer connection.Close()

		log.Printf("Writing...")
		_, err = connection.Write([]byte(request))
		if err != nil {
			return nil, err
		}

		log.Printf("Reading...")
		rawContents, err := ioutil.ReadAll(connection)
		if err != nil {
			return nil, err
		}

		pageContents := string(rawContents)
		return strings.Split(pageContents, "\r\n"), nil
	}

	Convey("Can manage connections", t, func() {
		lines, err := fetchPage(passingListener)
		So(err, ShouldBeNil)
		So(lines, ShouldContain, "Example-Header: IsSet")
		log.Printf("Done...")

		// Errors are hard to detect since events are async so
		// we just make sure that we don't get a full response back
		lines, _ = fetchPage(rejectingListener)
		So(lines, ShouldNotContain, "Example-Header: IsSet")
		So(len(lines), ShouldBeLessThan, 2)
	})
}
