package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/kgilpin/secretless-pg/conjur"
)

var AdminAPIKey = os.Getenv("CONJUR_AUTHN_API_KEY")
var Host = os.Getenv("TEST_PROXY_HOST")
var Port = os.Getenv("TEST_PROXY_PORT")

func psql(user string, environment []string) (string, error) {
	if Host == "" {
		Host = "proxy"
	}
	if Port == "" {
		Port = "5432"
	}

	args := []string{"-h", Host, "-p", Port, "-U", user, "-c", "select count(*) from conjur.test", "sslmode=disable dbname=postgres"}
	cmd := exec.Command("psql", args...)
	env := os.Environ()
	for _, v := range environment {
		env = append(env, v)
	}
	cmd.Env = env
	cmdOut, err := cmd.CombinedOutput()
	return string(cmdOut), err
}

func TestStaticPasswordLogin(t *testing.T) {
	log.Print("Provide a statically configured password")

	cmdOut, err := psql("alice", []string{"PGPASSWORD=alice"})

	if err != nil {
		t.Fatal(cmdOut)
	}

	if !strings.Contains(cmdOut, "1 row") {
		t.Fatalf("Expected to find '1 row' in : %s", cmdOut)
	}
}

func TestStaticPasswordLoginFailed(t *testing.T) {
	log.Print("Provide the wrong value for a statically configured password")

	cmdOut, err := psql("alice", []string{"PGPASSWORD=foobar"})

	if err == nil {
		t.Fatalf("Expected failed login : %s", cmdOut)
	}

	if !strings.Contains(cmdOut, "FATAL") {
		t.Fatalf("Expected to find 'FATAL' in : %s", cmdOut)
	}
	if !strings.Contains(cmdOut, "Login failed") {
		t.Fatalf("Expected to find 'Login failed' in : %s", cmdOut)
	}
}

func TestConjurLogin(t *testing.T) {
	log.Print("Provide a Conjur access token as the password")

	if AdminAPIKey == "" {
		t.Fatalf("CONJUR_AUTHN_API_KEY is missing")
	}

	var (
		adminToken *string
		userAPIKey *string
		userToken  *string
		err        error
	)

	if adminToken, err = conjur.Authenticate("admin", AdminAPIKey); err != nil {
		t.Fatalf("Failed to authenticate as 'admin' : %s", err)
	}
	if userAPIKey, err = conjur.RotateAPIKey("bob", *adminToken); err != nil {
		t.Fatalf("Failed to rotate API key of user 'bob'", err)
	}
	if userToken, err = conjur.Authenticate("bob", *userAPIKey); err != nil {
		t.Fatalf("Failed to authenticate as 'bob' : %s", err)
	}

	userToken64 := base64.StdEncoding.EncodeToString([]byte(*userToken))

	cmdOut, err := psql("bob", []string{fmt.Sprintf("PGPASSWORD=%s", userToken64)})

	if err != nil {
		t.Fatal(cmdOut)
	}

	if !strings.Contains(cmdOut, "1 row") {
		t.Fatalf("Expected to find '1 row' in : %s", cmdOut)
	}
}

func TestConjurUnauthorized(t *testing.T) {
	log.Print("Provide a Conjur access token for an unauthorized user")

	if AdminAPIKey == "" {
		t.Fatalf("CONJUR_AUTHN_API_KEY is missing")
	}

	var (
		adminToken *string
		userAPIKey *string
		userToken  *string
		err        error
	)

	if adminToken, err = conjur.Authenticate("admin", AdminAPIKey); err != nil {
		t.Fatalf("Failed to authenticate as 'admin' : %s", err)
	}
	if userAPIKey, err = conjur.RotateAPIKey("charles", *adminToken); err != nil {
		t.Fatalf("Failed to rotate API key of user 'charles'", err)
	}
	if userToken, err = conjur.Authenticate("charles", *userAPIKey); err != nil {
		t.Fatalf("Failed to authenticate as 'charles' : %s", err)
	}

	userToken64 := base64.StdEncoding.EncodeToString([]byte(*userToken))

	cmdOut, err := psql("charles", []string{fmt.Sprintf("PGPASSWORD=%s", userToken64)})

	if err == nil {
		t.Fatal(cmdOut)
	}

	if !strings.Contains(cmdOut, "FATAL") {
		t.Fatalf("Expected to find 'FATAL' in : %s", cmdOut)
	}
	if !strings.Contains(cmdOut, "Conjur authorization failed") {
		t.Fatalf("Expected to find 'Conjur authorization failed' in : %s", cmdOut)
	}
}
