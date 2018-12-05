package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Secretless will either be secretless:3306 (in Docker) or
// localhost:<mapped-port> (on the local machine)
func secretlessConfiguration() (host string, port int, options map[string]string) {
	// localhost:<mapped-port> (on the local machine)
	options = make(map[string]string)
	_, err := net.LookupIP("secretless")
	if err == nil {
		host = "secretless"
		port = 3306
	} else {
		host = "localhost"
		port = 13306
		options["--ssl-mode"] = "DISABLED"
	}
	return host, port, options
}

func mysql(host string, port int, user string, environment []string, options map[string]string, flags []string) (string, error) {
	var args []string
	if host != "" {
		args = append(args, "-h")
		args = append(args, host)
	}
	if port != 0 {
		args = append(args, "-P")
		args = append(args, fmt.Sprintf("%d", port))
	}
	if user != "" {
		args = append(args, "-u")
		args = append(args, user)
	}
	for k, v := range options {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	for _, v := range flags {
		args = append(args, v)
	}
	args = append(args, "-e")
	args = append(args, "select count(*) from testdb.test")

	log.Println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmd := exec.Command("mysql", args...)
	env := os.Environ()
	for _, v := range environment {
		env = append(env, v)
	}
	cmd.Env = env
	cmdOut, err := cmd.CombinedOutput()
	return string(cmdOut), err
}

func TestMySQLHandler(t *testing.T) {

	Convey("Connect over a UNIX socket", t, func() {

		Convey("With username, wrong password", func() {

			options :=map[string]string {
				"--socket": "sock/mysql.sock",
				"--password": "wrongpassword",
			}

			cmdOut, err := mysql("", 0, "testuser", []string{}, options, []string{})

			So(err, ShouldBeNil)
			So(cmdOut, ShouldContainSubstring, "2")
		})

		Convey("With wrong username, wrong password", func() {

			options :=map[string]string {
				"--socket": "sock/mysql.sock",
				"--password": "wrongpassword",
			}

			cmdOut, err := mysql("", 0, "wrongusername", []string{}, options, []string{})

			So(err, ShouldBeNil)
			So(cmdOut, ShouldContainSubstring, "2")
		})

		Convey("With empty username, empty password", func() {

			options :=map[string]string {
				"--socket": "sock/mysql.sock",
				"--password": "",
			}

			cmdOut, err := mysql("", 0, "", []string{}, options, []string{})

			So(err, ShouldBeNil)
			So(cmdOut, ShouldContainSubstring, "2")
		})
	})

	Convey("Connect over TCP", t, func() {

		Convey("No TLS Upstream, TLS Downstream and sslmode default", func() {

			Convey("With username, wrong password", func() {

				host, port, options := secretlessConfiguration()
				options["--password"] = "wrongpassword"

				cmdOut, err := mysql(host, port, "testuser", []string{}, options, []string{})

				So(err, ShouldBeNil)
				So(cmdOut, ShouldContainSubstring, "2")
			})

			Convey("With wrong username, wrong password", func() {

				host, port, options := secretlessConfiguration()
				options["--password"] = "wrongpassword"

				cmdOut, err := mysql(host, port, "notatestuser", []string{}, options, []string{})

				So(err, ShouldBeNil)
				So(cmdOut, ShouldContainSubstring, "2")
			})

			Convey("With empty username, empty password", func() {

				host, port, options := secretlessConfiguration()
				options["--password"] = ""

				cmdOut, err := mysql(host, port, "", []string{}, options, []string{})

				So(err, ShouldBeNil)
				So(cmdOut, ShouldContainSubstring, "2")
			})
		})

		Convey("Upstream SSL", func() {

			host, port, options := secretlessConfiguration()
			options["--password"] = ""
			flags := []string{"--ssl"}

			_, err := mysql(host, port, "", []string{}, options, flags)

			So(err, ShouldBeError)
		})

		Convey("sslmode default", func() {
			Convey("Connect over TCP to server with TLS support", func() {

				options := make(map[string]string)
				host := "secretless"
				port := 3306
				options["--password"] = ""

				cmdOut, err := mysql(host, port, "", []string{}, options, []string{})

				So(err, ShouldBeNil)
				So(cmdOut, ShouldContainSubstring, "2")
			})

			Convey("Connect over TCP to server without TLS support", func() {

				options := make(map[string]string)
				host := "secretless"
				port := 4306
				options["--password"] = ""

				cmdOut, err := mysql(host, port, "", []string{}, options, []string{})

				So(err, ShouldNotBeNil)
				So(cmdOut, ShouldContainSubstring, "ERROR 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it")
			})
		})
	})
}
