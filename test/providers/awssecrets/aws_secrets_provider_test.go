package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

	awsinternal "github.com/cyberark/secretless-broker/internal/plugin/connectors/http/aws"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
	"github.com/cyberark/secretless-broker/internal/providers/awssecrets"
)

// Works uses a mock server to capture the request from the provider and compare
// the authorization header with a manually generated one using the same credentials.
// Assertions are also made against the request payload and response payload.
func Works(
	t *testing.T,
	resignCreds *credentials.Credentials,
	setupProviderAuth func() func(),
) {
	var originalAuthHeaders []string
	var reSignedAuthHeaders []string
	var secretInputs []secretsmanager.GetSecretValueInput

	// secretStringNotBinary determines the type of secret output payload returned by the
	// mock server, String or Binary. Default is String.
	secretStringNotBinary := true

	// TODO: This should really be standalone test fixture, that maintains a map of
	// 	request-secret-id to the latest request details and response details, or some
	// 	other thing that is unique to a request. This will allow for arbitrarily complex
	//	and robust testing using this mock server.
	// NOTE: This server can and has been run standalone and consumed by the AWS CLI to
	// confirm that it behaves as expected. Automating such a test wouldn't hurt.

	// Mock server for fetching secrets. Secret value in the response is always the value
	// "[request-secret-id]-value"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Parse and capture the request secretInput
		var secretInput secretsmanager.GetSecretValueInput
		err = json.Unmarshal(reqBytes, &secretInput)
		if err != nil {
			t.Fatal(err)
		}
		secretInputs = append(secretInputs, secretInput)

		// Parse request for AWS signed request metadata
		reqMeta, err := awsinternal.NewRequestMetadata(r)
		if err != nil {
			t.Fatal(err)
		}

		// Capture original signed request header
		originalAuthHeaders = append(originalAuthHeaders, r.Header.Get("Authorization"))

		// Remove all non-signed headers
		signedHeaders := map[string]bool{}
		for _, name := range reqMeta.SignedHeaders {
			signedHeaders[strings.ToLower(name)] = true
		}
		for name := range r.Header {
			if ok := signedHeaders[strings.ToLower(name)]; ok {
				continue
			}

			r.Header.Del(name)
		}

		// Re-sign request
		_, err = v4.NewSigner(resignCreds).Sign(
			r,
			bytes.NewReader(reqBytes),
			reqMeta.ServiceName,
			reqMeta.Region,
			reqMeta.Date,
		)
		if err != nil {
			t.Fatal(err)
		}

		// Capture resigned request header
		reSignedAuthHeaders = append(reSignedAuthHeaders, r.Header.Get("Authorization"))

		// Craft response. The first response is secret string, then all subsequent
		// responses are secret binary
		var secretOutput secretsmanager.GetSecretValueOutput
		secretValue := aws.StringValue(secretInput.SecretId) + "-value"
		if secretStringNotBinary {
			secretOutput.SetSecretString(secretValue)
		} else {
			secretOutput.SetSecretBinary([]byte(secretValue))
		}

		// Write response
		responseBytes, err := json.Marshal(secretOutput)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBytes)
	}))
	defer server.Close()

	// Setup authentication for provider
	cleanup := setupProviderAuth()
	defer cleanup()

	// Create provider (with custom endpoint)
	p, err := awssecrets.NewProvider(plugin_v1.ProviderOptions{Name: "aws"}, aws.Config{
		Endpoint: aws.String(server.URL),
	})
	So(err, ShouldBeNil)

	// Make 2 attempts to get values from provider
	secretStringNotBinary = true
	secretValue0, err := p.GetValue("meow-id-0")
	So(err, ShouldBeNil)

	secretStringNotBinary = false
	secretValue1, err := p.GetValue("meow-id-1")
	So(err, ShouldBeNil)

	// Ensure 2 attempts are recorded
	So(reSignedAuthHeaders, ShouldHaveLength, 2)

	// Assert that the auth header sent by the Provider matches the manually
	// signed one
	So(reSignedAuthHeaders[0], ShouldEqual, originalAuthHeaders[0])
	So(reSignedAuthHeaders[1], ShouldEqual, originalAuthHeaders[1])
	// Assert on the requested secret id
	So(aws.StringValue(secretInputs[0].SecretId), ShouldEqual, "meow-id-0")
	So(aws.StringValue(secretInputs[1].SecretId), ShouldEqual, "meow-id-1")
	// Assert on the response secret values
	So(string(secretValue0), ShouldEqual, "meow-id-0-value")
	So(string(secretValue1), ShouldEqual, "meow-id-1-value")
}

func TestAWSSecrets_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider

	name := "aws"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	Convey("Can create the AWS Secrets provider", t, func() {
		provider, err = providers.ProviderFactories[name](options)
		So(err, ShouldBeNil)
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, "aws")
	})


	Convey("Fetches credentials using AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY", t, func() {
		Works(
			t,
			credentials.NewStaticCredentials(
				"xyz",
				"abc",
				"",
			),
			func() func() {
				_ = os.Setenv("AWS_ACCESS_KEY_ID", "xyz")
				_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "abc")
				return func() {
					_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
					_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				}
			},
		)
	})

	Convey("Fetches credentials using AWS_SESSION_TOKEN", t, func() {
		Works(
			t,
			credentials.NewStaticCredentials(
				"abc",
				"meow",
				"moo",
			),
			func() func() {
				_ = os.Setenv("AWS_ACCESS_KEY_ID", "abc")
				_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "meow")
				_ = os.Setenv("AWS_SESSION_TOKEN", "moo")
				return func() {
					_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
					_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
					_ = os.Unsetenv("AWS_SESSION_TOKEN")
				}
			},
		)
	})
}
