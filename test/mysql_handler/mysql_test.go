package main

import (
        "fmt"
        "log"
        "os"
        "os/exec"
        "strings"
        "testing"

        . "github.com/smartystreets/goconvey/convey"
)

func mysql(host string, port int, user string, environment []string, options map[string]string) (string, error) {
	args := []string{}
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
		if v != "" {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		} else {
			args = append(args, k)
		}
	}
	args = append(args, "--password=wrongpassword")
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

	Convey("Connect over a UNIX socket w/username", t, func() {

		options := make(map[string]string)
		options["--socket"] = "run/mysql/mysql.sock"

		cmdOut, err := mysql("", 0, "testuser", []string{}, options)

		fmt.Printf("err: %v\n", err)
		fmt.Printf("cmdOut: %v\n", cmdOut)

		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "2")
	})

	Convey("Connect over a UNIX socket w/invalid username", t, func() {

		options := make(map[string]string)
                options["--socket"] = "run/mysql/mysql.sock"

                cmdOut, err := mysql("", 0, "dummy", []string{}, options)

                fmt.Printf("err: %v\n", err)
                fmt.Printf("cmdOut: %v\n", cmdOut)

                So(err, ShouldBeNil)
                So(cmdOut, ShouldContainSubstring, "2")

	})

	Convey("Connect over a UNIX socket w/o username", t, func() {

                options := make(map[string]string)
                options["--socket"] = "run/mysql/mysql.sock"

                cmdOut, err := mysql("", 0, "", []string{}, options)

                fmt.Printf("err: %v\n", err)
                fmt.Printf("cmdOut: %v\n", cmdOut)

                So(err, ShouldBeNil)
                So(cmdOut, ShouldContainSubstring, "2")

        })

/*
	// This is not currently implemented
	Convey("Connect over TCP", t, func() {
		// Secretless will either be secretless:3306 (in Docker) or
		// localhost:<mapped-port> (on the local machine)
		var host string
		var port int
		_, err := net.LookupIP("mysql")
		if err == nil {
			host = "secretless"
			port = 3306
		} else {
			host = "localhost"
			port = 13306
		}

		cmdOut, err := mysql(host, port, "", []string{})

		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "1 row")
	})
*/
}
