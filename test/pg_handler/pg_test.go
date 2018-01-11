package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var AdminAPIKey = os.Getenv("CONJUR_AUTHN_API_KEY")
var Host = os.Getenv("TEST_PROXY_HOST")

func psql(host string, port int, user string, environment []string) (string, error) {
	if host == "" {
		if Host != "" {
			host = Host
		} else {
			host = "secretless_test"
		}
	}

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
	args = append(args, "dbname=postgres")

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
func TestUnixSocketConnection(t *testing.T) {
	log.Print("Connect via Unix socket without authentication")

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmdOut, err := psql(fmt.Sprintf("%s/run/postgresql", cwd), 0, "", []string{})

	if err != nil {
		t.Fatal(cmdOut)
	}

	if !strings.Contains(cmdOut, "1 row") {
		t.Fatalf("Expected to find '1 row' in : %s", cmdOut)
	}
}
