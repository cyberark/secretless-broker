package troubleshooting

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type client func (url, proxy string) (string, error)

func Code(text string) string {
	return "```\n" + text + "\n```"
}

func Indent(text, indent string) string {
	return _Indent(Code(strings.TrimSpace(text)), indent)
}

// indents a block of text with an indent string
func _Indent(text, indent string) string {
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

func httpCurl(url, proxy string) (string, error)  {
	args := []string{
		//"-k",
		"-sS",
		"-v",
		"-X", "GET",
		url,
	}

	cmd := exec.Command(
		"curl",
		args...,
	)

	proxyVarName := "http_proxy"
	if strings.HasPrefix(url, "https") {
		proxyVarName = "https_proxy"
	}

	cmd.Env = append(cmd.Env, proxyVarName+"="+proxy)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("%s", string(out))
	}

	return string(out), nil
}

func httpGo(url, proxy string) (string, error)  {
	res, err := proxyGet(
		url,
		proxy,
	)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
Status:
%s

Body: 
%s`, res.Status, string(body)), nil
}

func TestHTTPS(t *testing.T) {
	_, logs, err1 := troubleShooting(t, "https://httpbin.org/anything", httpCurl)
	if !assert.Error(t, err1) {
		return
	}

	_, _, err2 := troubleShooting(t, "https://httpbin.org/anything", httpGo)
	if !assert.Error(t, err2) {
		return
	}

	res := fmt.Sprintf(`
### Connecting to Secretless as an HTTPS proxy
#### Symptoms
- You see a CONNECT request in your Secretless logs that looks something like:

%s

- Sample client log output messages
  
  curl:
%s

  Go client:
%s

#### Known Causes
This type of error occurs when the client attempts to use Secretless as an HTTPS proxy. 
Secretless can only act as an HTTP proxy.
This error can happen for a few reasons:
1. An explicit attempt to use Secretless as an HTTPS proxy
1. Providing the client an HTTPS target when intending to proxy the connection through Secretless might result in the client attempting to use Secretless as an HTTPS proxy, as is the case with Go's standard library HTTP client.

#### Resolution
- Ensure that target of your request is HTTP only e.g. http://httpbin.org.
- Secretless does not support HTTPS between the client and Secretless, it does support it between Secretless and the target. Do not use Secretless as an HTTPS proxy. 
- To make connection between Secretless and the target an HTTPS connection you must set "forceSSL: true" on the Secretless service connector config. 

`, Indent(logs.String(), "  "), Indent(err1.Error(), "  "), Indent(err2.Error(), "  "))

	writeToTroubleshooting(res)
}

func TestSelfSigned(t *testing.T) {
	out1, logs, err := troubleShooting(t, "http://self-signed.badssl.com", httpCurl)
	if !assert.NoError(t, err) {
		return
	}

	out2, _, err := troubleShooting(t, "http://wrong.host.badssl.com/", httpGo)
	if !assert.NoError(t, err) {
		return
	}

	res := fmt.Sprintf(`
### HTTPS certificate verification failure when forceSSL is set
#### Symptoms
- You see an x509 certificate error in your Secretless logs that looks something like:

%s

- Sample client log output messages
  
  curl:
%s

  Go client:
%s

#### Known Causes
This type of error occurs when the client attempts to connect to a target with a self-signed certificate, and there is some failure on verification. Secretless verifies all HTTPS connections to the target.

There are several reasons why verification might fail including:
1. The signer of the target's certificate is not a trusted CA
1. The target's certificate is expired or is not yet valid
1. The target's certificate is not valid for the host

#### Resolution

This type of error can be broken into 2 categories.
1. The signer of the target's certificate is not a trusted CA
1. The rest

For the rest (2), you must ensure that the target's certificate is valid for the target. 

For (1) you will need to ensure that Secretless is aware of the root certificate authority (CA) it should use to verify the server certificates when proxying requests. To do this, ensure that the environment variable **SECRETLESS_HTTP_CA_BUNDLE** is set. **SECRETLESS_HTTP_CA_BUNDLE** is a path to the bundle of CA certificates that are appended to the certificate pool that Secretless uses for server certificate verification of all HTTP service connectors.
`, Indent(logs.String(), "  "), Indent(out1, "  "), Indent(out2, "  "))


	writeToTroubleshooting(res)
}

func writeToTroubleshooting(out string) {
	f, err := os.OpenFile("troubleshooting.md",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString(out); err != nil {
		panic(err)
	}
}
func troubleShooting(t *testing.T, url string, client2 client) (string, bytes.Buffer, error) {
	var buf bytes.Buffer
	// Create in-process proxy service
	proxyService, err := newInProcessProxyService(
		[]byte(`
credentialPatterns:
  username: '[^:]+'
headers:
  Authorization: ' Basic {{ printf "%s:%s" .username .password | base64 }}'
authenticateURLsMatching:
  - ^http
forceSSL: true
`),
		map[string][]byte{},
		&buf,
		)
	if !assert.NoError(t, err) {
		return "", buf, nil
	}

	// Ensure the proxy service is stopped
	defer proxyService.Stop()
	// Start the proxyService service. Note
	proxyService.Start()

	// Avoid all the unnecessary initial logs
	buf.Reset();

	proxyURL := "http://" + proxyService.host + ":" + proxyService.port

	// Make the client request to the proxy service
	out, err := client2(url, proxyURL)
	return out, buf, err
}
