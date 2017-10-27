package conjur

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

type Method struct {
	Method      string
	Path        string
	Body        io.Reader
	AccessToken *string
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

	if method.AccessToken != nil {
		req.Header.Set(
			"Authorization",
			fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString([]byte(*method.AccessToken))),
		)
	}

	return HttpClient.Do(req)
}

func Authenticate(username string, apiKey string) (*string, error) {
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
			return &response, nil
		}
	} else {
		return nil, fmt.Errorf("Authentication failed for user '%s' : %d", username, resp.StatusCode)
	}
}

func CheckPermission(resource string, token string) (bool, error) {
	resp, err := Request(&Method{
		Path:        fmt.Sprintf("/resources/%s?check=true&privilege=execute", PathOfId(resource)),
		Method:      "GET",
		AccessToken: &token,
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

func RotateAPIKey(username string, token string) (*string, error) {
	resp, err := Request(&Method{
		Path:        fmt.Sprintf("/authn/dev/api_key?role=dev:user:%s", username),
		Method:      "PUT",
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

func Secret(resource, token string) (string, error) {
	resp, err := Request(&Method{
		Path:        fmt.Sprintf("/secrets/%s", PathOfId(resource)),
		Method:      "GET",
		AccessToken: &token,
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
