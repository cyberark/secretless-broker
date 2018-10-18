package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func psql(host string, port int, user string, environment []string) (string, error) {
	args := []string{"-h", host}
	if port != 0 {
		args = append(args, "-p")
		args = append(args, fmt.Sprintf("%d", port))
	}
	if user != "" {
		args = append(args, "-U")
		args = append(args, user)
	}
	args = append(args, "-c")
	args = append(args, "select count(*) from test.test")
	args = append(args, "dbname=postgres sslmode=disable")

	log.Println(strings.Join(append([]string{"psql"}, args...), " "))

	cmd := exec.Command("psql", args...)
	env := os.Environ()
	for _, v := range environment {
		env = append(env, v)
	}
	cmd.Env = env
	cmdOut, err := cmd.CombinedOutput()
	return string(cmdOut), err
}
func TestPGHandler(t *testing.T) {

	Convey("Connect over a Unix socket", t, func() {
		cmdOut, err := psql("/sock", 0, "", []string{})
		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "1 row")
	})

	Convey("Connect over TCP", t, func() {
		// Secretless will either be secretless:5432 (in Docker) or
		// localhost:<mapped-port> (on the local machine)
		var host string
		var port int
		_, err := net.LookupIP("pg")
		if err == nil {
			host = "secretless"
			port = 15432
		} else {
			host = "localhost"
			port, err = strconv.Atoi(os.Getenv("SECRETLESS_PORT"))
			if err != nil {
				t.Error(err)
			}
		}

		cmdOut, err := psql(host, port, "", []string{})

		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "1 row")
	})

	Convey("Connect over TCP with TLS downstream", t, func() {
		_, err := net.LookupIP("pg")
		if err != nil {
			t.Error(err)
		}

		host := "secretless"
		port := 25432

		cmdOut, err := psql(host, port, "", []string{})

		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "1 row")
	})
}
