package conjur

import (
  "encoding/base64"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "strings"
  "time"

  "github.com/cyberark/conjur-api-go/conjurapi"
)

type AccessToken struct {
  Token      string
  UseDefault bool
}

type Method struct {
  Method      string
  Path        string
  Body        io.Reader
  AccessToken AccessToken
}

var Config = conjurapi.LoadConfig()
var HttpClient = &http.Client{Timeout: time.Second * 10}

func PathOfId(id string) string {
  tokens := strings.SplitN(id, ":", 3)
  return strings.Join(tokens, "/")
}

func Request(method *Method) (*http.Response, error) {
  req, err := http.NewRequest(
    method.Method,
    fmt.Sprintf("%s%s", Config.ApplianceURL, method.Path),
    method.Body,
  )
  if err != nil {
    return nil, err
  }

  if method.AccessToken.UseDefault && method.AccessToken.Token == "" {
  	if token, err := DefaultAccessToken(); err != nil {
  		return nil, err
  	} else {
	    method.AccessToken = *token
  	}
  }

  if method.AccessToken.Token != "" {
    req.Header.Set(
      "Authorization",
      fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString([]byte(method.AccessToken.Token))),
    )
  }

  return HttpClient.Do(req)
}

func Username() (string, error) {
  if result := os.Getenv("CONJUR_AUTHN_LOGIN"); result == "" {
    return "", fmt.Errorf("CONJUR_AUTHN_LOGIN is not specified")
  } else {
    return result, nil
  }
}

func APIKey() (apiKey string, err error) {
  if apiKey = os.Getenv("CONJUR_AUTHN_API_KEY"); apiKey == "" {
    return "", fmt.Errorf("CONJUR_AUTHN_API_KEY is not specified")
  } else {
    return
  }
}

func DefaultAccessToken() (*AccessToken, error) {
  var err error
  var username, apiKey string

  if username, err = Username(); err != nil {
    return nil, fmt.Errorf("Conjur username is not available")
  }
  if apiKey, err = APIKey(); err != nil {
    return nil, fmt.Errorf("Conjur API key is not available")
  }

  if token, err := Authenticate(username, apiKey); err != nil {
    return nil, err
  } else {
    return token, nil
  }
}

func Authenticate(username string, apiKey string) (*AccessToken, error) {
  resp, err := Request(&Method{
    Path:   fmt.Sprintf("/authn/dev/%s/authenticate", url.PathEscape(username)),
    Method: "POST",
    Body:   strings.NewReader(apiKey),
  })
  if err != nil {
    return nil, err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    if responseBytes, err := ioutil.ReadAll(resp.Body); err != nil {
      return nil, err
    } else {
      response := string(responseBytes)
      return &AccessToken{Token: response}, nil
    }
  } else {
    return nil, fmt.Errorf("Authentication failed for user '%s' : %d", username, resp.StatusCode)
  }
}

func CheckPermission(resource string, token AccessToken) (bool, error) {
  token.UseDefault = true

  resp, err := Request(&Method{
    Path:        fmt.Sprintf("/resources/%s?check=true&privilege=execute", PathOfId(resource)),
    Method:      "GET",
    AccessToken: token,
  })

  if err != nil {
    return false, err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    log.Printf("User is authorized to 'execute' %s", resource)
    return true, nil
  } else if resp.StatusCode == 404 || resp.StatusCode == 403 {
    return false, nil
  } else {
    return false, fmt.Errorf("Permission check failed with HTTP status %d", resp.StatusCode)
  }
}

func RotateAPIKey(username string, token AccessToken) (string, error) {
  token.UseDefault = true

  resp, err := Request(&Method{
    Path:        fmt.Sprintf("/authn/dev/api_key?role=dev:user:%s", username),
    Method:      "PUT",
    AccessToken: token,
  })
  if err != nil {
    return "", err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    if responseBytes, err := ioutil.ReadAll(resp.Body); err != nil {
      return "", err
    } else {
      return string(responseBytes), nil
    }
  } else {
    return "", fmt.Errorf("Failed to rotate API key for user '%s' : %d", username, resp.StatusCode)
  }
}

func Secret(resource string, token AccessToken) (string, error) {
  token.UseDefault = true

  resp, err := Request(&Method{
    Path:        fmt.Sprintf("/secrets/%s", PathOfId(resource)),
    Method:      "GET",
    AccessToken: token,
  })
  if err != nil {
    return "", err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    if responseBytes, err := ioutil.ReadAll(resp.Body); err != nil {
      return "", err
    } else {
      response := string(responseBytes)
      return response, nil
    }
  } else {
    return "", fmt.Errorf("Failed to fetch secret with HTTP status %d", resp.StatusCode)
  }
}
