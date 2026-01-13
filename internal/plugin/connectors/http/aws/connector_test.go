package aws

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/log"
)

const authzHeader = "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-date, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024"

type connectTestCase struct {
	description     string
	url             string
	headers         map[string][]string
	credentialsByID map[string][]byte
	assert          func(t *testing.T, beforeR *http.Request, afterR *http.Request, err error)
}

const connectTestCaseBodyContents = "xyz"
const connectTestCaseBodySHA256 = "3608bca1e44ea6c4d268eb6db02260269892c0b42b86bbf1e77a6fa16c3c9282"

func (c connectTestCase) Run(t *testing.T) {
	t.Run(c.description, func(t *testing.T) {
		// Arrange: test server for endpoint "discovery"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer ts.Close()

		// Simulate endpoint discovery via non-deprecated env var for S3.
		// Only set it for the test case that uses the placeholder URL.
		if strings.Contains(c.url, "secretless.empty") {
			t.Setenv("AWS_ENDPOINT_URL_S3", ts.URL)
		}

		// Prepare request body so X-Amz-Content-Sha256 is deterministic.
		var buf bytes.Buffer
		buf.Write([]byte(connectTestCaseBodyContents))

		beforeR, _ := http.NewRequest("PUT", c.url, &buf)
		beforeR.Header.Set("x-amz-content-sha256", "this-will-be-generated-at-signing")

		// Unsigned headers
		beforeR.Header.Set("Unsigned-Header-1", "Unsigned-Header-1-Value")
		beforeR.Header.Set("Unsigned-Header-2", "Unsigned-Header-2-Value")

		for key, values := range c.headers {
			for _, value := range values {
				beforeR.Header.Add(key, value)
			}
		}

		// Create a clone of the original request. We need the original for comparisons
		// during assertion.
		afterR := beforeR.Clone(context.Background())
		// This is needed because the cloning mechanism for the body isn't deep.
		// See https://github.com/golang/go/issues/36095
		afterR.Body, _ = beforeR.GetBody()

		// Call Connect method using the clone of the original request. Some of our assertions
		// will be based on the comparison of the original and the clone, since Connect will
		// potentially mutate the request passed to it.
		conn := Connector{logger: log.NewWithOptions(io.Discard, "", false)}
		err := conn.Connect(afterR, c.credentialsByID)

		c.assert(t, beforeR, afterR, err)

		// For visibility when debugging failures:
		_ = ts // keep ts in scope to avoid accidental removal by IDE refactors
	})
}

var testCases = []connectTestCase{
	{
		description:     "no signing",
		url:             "http://meow.moo",
		headers:         nil,
		credentialsByID: nil,
		assert: func(t *testing.T, beforeR *http.Request, afterR *http.Request, err error) {
			assert.NoError(t, err)

			beforeRDump, err := httputil.DumpRequest(beforeR, true)
			assert.NoError(t, err)
			afterRDump, err := httputil.DumpRequest(afterR, true)
			assert.NoError(t, err)

			// The request should remain the same before and after because there is no
			// initial-signing to override.
			assert.Equal(t, string(beforeRDump), string(afterRDump))
		},
	},
	{
		description: "signing without endpoint discovery",
		url:         "http://meow.moo",
		headers: map[string][]string{
			"Authorization": {authzHeader},
			"X-Amz-Date":    {"20210102T150405Z"},
		},
		credentialsByID: map[string][]byte{
			"accessKeyId":     []byte("accessKeyIdValue"),
			"secretAccessKey": []byte("secretAccessKeyValue"),
		},
		assert: func(t *testing.T, beforeR *http.Request, afterR *http.Request, err error) {
			assert.NoError(t, err)

			// The request URL should remain the same because endpoint discovery is not
			// being used
			assert.Equal(t, beforeR.URL, afterR.URL)

			// The Authorization should be modified and should use the injected credentials
			assert.NotEqual(t, beforeR.Header.Get("Authorization"), afterR.Header.Get("Authorization"))
			assert.Equal(t,
				// The expected Authorization header value has been manually calculated
				"AWS4-HMAC-SHA256 Credential=accessKeyIdValue/20210102/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=44890edc510facaf7f55bec0cb8eb04a1444690d265858fdab15123b489d5d7b",
				afterR.Header.Get("Authorization"),
			)

			assertOnHeadersAfterSigning(t, afterR)
		},
	},
	{
		description: "signing with endpoint discovery",
		url:         "http://secretless.empty",
		headers: map[string][]string{
			"Authorization": {authzHeader},
			"X-Amz-Date":    {"20210102T150405Z"},
		},
		credentialsByID: map[string][]byte{
			"accessKeyId":     []byte("accessKeyIdValue"),
			"secretAccessKey": []byte("secretAccessKeyValue"),
		},
		assert: func(t *testing.T, beforeR *http.Request, afterR *http.Request, err error) {
			assert.NoError(t, err)

			// The request URL is changed to one determined by endpoint discovery
			assert.NotEqual(t, beforeR.URL.Host, afterR.URL.Host)
			assert.NotEmpty(t, afterR.URL.Host)

			// Authorization is re-signed with injected creds; validate parts, not exact hex.
			got := afterR.Header.Get("Authorization")
			assert.True(t,
				strings.HasPrefix(got, "AWS4-HMAC-SHA256"),
				"Authorization must start with scheme",
			)
			// Credential scope
			credScope := "Credential=accessKeyIdValue/20210102/us-east-1/s3/aws4_request"
			assert.Contains(t, got, credScope, "Credential scope must match expected")
			// SignedHeaders
			assert.Contains(t, got, "SignedHeaders=host;x-amz-content-sha256;x-amz-date")
			// Signature length (64 hex chars)
			sig := extractSignature(got)
			assert.Equal(t, 64, len(sig), "Signature length must be 64 hex characters")
			_, toHexError := hex.DecodeString(sig)
			assert.NoError(t, toHexError, "Signature must be hex")

			assertOnHeadersAfterSigning(t, afterR)
		},
	},
	{
		description: "missing credentials for signing",
		url:         "http://meow.moo",
		headers: map[string][]string{
			"Authorization": {authzHeader},
			"X-Amz-Date":    {"20060102T150405Z"},
		},
		assert: func(t *testing.T, beforeR *http.Request, afterR *http.Request, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "missing required AWS credentials")
		},
	},
}

func assertOnHeadersAfterSigning(t *testing.T, afterR *http.Request) {
	// X-Amz-Content-Sha256 is always recalculated
	assert.Equal(t, connectTestCaseBodySHA256, afterR.Header.Get("X-Amz-Content-Sha256"))

	// Unsigned headers remain unchanged
	assert.Equal(t, "Unsigned-Header-1-Value", afterR.Header.Get("Unsigned-Header-1"))
	assert.Equal(t, "Unsigned-Header-2-Value", afterR.Header.Get("Unsigned-Header-2"))
}

func TestConnector_Connect(t *testing.T) {
	for _, testCase := range testCases {
		testCase.Run(t)
	}
}

// Helpers for Authorization header validation
func extractSignature(authz string) string {
	parts := strings.Split(authz, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "Signature=") {
			return strings.TrimPrefix(p, "Signature=")
		}
	}
	return ""
}
