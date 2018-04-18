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

func mysql(host string, port int, user string, environment []string) (string, error) {
	args := []string{"-h", host}
	if port != 0 {
		args = append(args, "-P")
		args = append(args, fmt.Sprintf("%d", port))
	}
	if user != "" {
		args = append(args, "-u")
		args = append(args, user)
	}
	args = append(args, "-e")
	args = append(args, "select count(*) from test.test")

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
		cwd, err := os.Getwd()
		if err != nil
			panic(err)
		}

		cmdOut, err := mysql(fmt.Sprintf("%s/run/mysql", cwd), 0, "", []string{})

		So(err, ShouldBeNil)
		So(cmdOut, ShouldContainSubstring, "1 row")
	})

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
}
