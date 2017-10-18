package main

import (
  "log"
  "fmt"
  "time"
  "testing"
  "os"
  "os/exec"
  "io"
  "io/ioutil"
  "strings"
  "net/http"
  "encoding/base64"

  "github.com/cyberark/conjur-api-go/conjurapi"
)

type ConjurMethod struct {
  Method string
  Path   string
  Body   io.Reader
  AccessToken *string
}

var ConjurConfig = conjurapi.LoadConfig()
var AdminAPIKey  = os.Getenv("CONJUR_AUTHN_API_KEY")

func psql(user string, environment []string) (string, error) {
  args := []string{"-h", "proxy", "-p", "5432", "-U", user, "-c", "select count(*) from conjur.test", "sslmode=disable dbname=postgres"}
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

  cmdOut, err := psql("alice", []string{ "PGPASSWORD=alice" })

  if err != nil {
    t.Fatal(cmdOut)
  }

  if !strings.Contains(cmdOut, "1 row") {
    t.Fatalf("Expected to find '1 row' in : %s", cmdOut)
  }
}

func TestStaticPasswordLoginFailed(t *testing.T) {
  log.Print("Provide the wrong value for a statically configured password")

  cmdOut, err := psql("alice", []string{ "PGPASSWORD=foobar" })

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

func conjurRequest(method *ConjurMethod) (*http.Response, error) {
  httpClient := &http.Client{Timeout: time.Second * 10}

  req, err := http.NewRequest(
    method.Method,
    fmt.Sprintf("%s%s", ConjurConfig.ApplianceURL, method.Path),
    method.Body,
  )
  if ( err != nil ) {
    return nil, err
  }

  if method.AccessToken != nil {
    req.Header.Set(
      "Authorization",
      fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString([]byte(*method.AccessToken))),
    )
  }
  
  return httpClient.Do(req)
}

func conjurAuthenticateUser(username string, apiKey string) (*string, error) {
  resp, err := conjurRequest(&ConjurMethod{
    Path: fmt.Sprintf("/authn/dev/%s/authenticate", username),
    Method: "POST",
    Body: strings.NewReader(apiKey),
  })
  if err != nil {
    return nil, err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    if responseBytes, err := ioutil.ReadAll(resp.Body); err != nil {
      return nil, err
    } else {
      response := string(responseBytes)
      return &response, nil
    }    
  } else {
    return nil, fmt.Errorf("Authentication failed for user '%s' : %d", username, resp.StatusCode)
  }  
}

func conjurRotateAPIKey(username string, token string) (*string, error) {
  resp, err := conjurRequest(&ConjurMethod{
    Path: fmt.Sprintf("/authn/dev/api_key?role=dev:user:%s", username),
    Method: "PUT",
    AccessToken: &token,
  })
  if err != nil {
    return nil, err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    if responseBytes, err := ioutil.ReadAll(resp.Body); err != nil {
      return nil, err
    } else {
      response := string(responseBytes)
      return &response, nil
    }
  } else {
    return nil, fmt.Errorf("Failed to rotate API key for user '%s' : %d", username, resp.StatusCode)
  }  
}

func TestConjurLogin(t *testing.T) {
  log.Print("Provide a Conjur access token as the password")

  if AdminAPIKey == "" {
    t.Fatalf("CONJUR_AUTHN_API_KEY is missing")
  }

  var (
    adminToken *string
    userAPIKey  *string
    userToken   *string
    err   error
  )

  if adminToken, err = conjurAuthenticateUser("admin", AdminAPIKey); err != nil {
    t.Fatalf("Failed to authenticate as 'admin' : %s", err)
  }
  if userAPIKey, err = conjurRotateAPIKey("bob", *adminToken); err != nil {
    t.Fatalf("Failed to rotate API key of user 'bob'", err)
  }
  if userToken, err = conjurAuthenticateUser("bob", *userAPIKey); err != nil {
    t.Fatalf("Failed to authenticate as 'bob' : %s", err)
  }

  userToken64 := base64.StdEncoding.EncodeToString([]byte(*userToken))

  cmdOut, err := psql("bob", []string{ fmt.Sprintf("PGPASSWORD=%s", userToken64) })

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
    userAPIKey  *string
    userToken   *string
    err   error
  )

  if adminToken, err = conjurAuthenticateUser("admin", AdminAPIKey); err != nil {
    t.Fatalf("Failed to authenticate as 'admin' : %s", err)
  }
  if userAPIKey, err = conjurRotateAPIKey("charles", *adminToken); err != nil {
    t.Fatalf("Failed to rotate API key of user 'charles'", err)
  }
  if userToken, err = conjurAuthenticateUser("charles", *userAPIKey); err != nil {
    t.Fatalf("Failed to authenticate as 'charles' : %s", err)
  }

  userToken64 := base64.StdEncoding.EncodeToString([]byte(*userToken))

  cmdOut, err := psql("charles", []string{ fmt.Sprintf("PGPASSWORD=%s", userToken64) })

  if err == nil {
    t.Fatal(cmdOut)
  }

  if !strings.Contains(cmdOut, "FATAL") {
    t.Fatalf("Expected to find 'FATAL' in : %s", cmdOut)
  }
  if !strings.Contains(cmdOut, "Conjur authentication failed") {
    t.Fatalf("Expected to find 'Conjur authentication failed' in : %s", cmdOut)
  }
}
