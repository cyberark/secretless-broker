package aws

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
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
		// Create original request before connect is called

		// Use a request body with standard contents for ease of calculating
		// x-amz-content-sha256
		var buf bytes.Buffer
		buf.Write([]byte(connectTestCaseBodyContents))

		beforeR, _ := http.NewRequest("PUT", c.url, &buf)
		beforeR.Header.Set(
			"x-amz-content-sha256",
			"this-will-be-generated-at-signing",
		)

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
		connector := Connector{logger: log.NewWithOptions(ioutil.Discard, "", false)}
		err := connector.Connect(afterR, c.credentialsByID)

		// Make assertions
		c.assert(t, beforeR, afterR, err)
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
			assert.Equal(
				t,
				string(beforeRDump),
				string(afterRDump),
			)
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
			assert.Equal(
				t,
				beforeR.URL,
				afterR.URL,
			)
			// The Authorization should be modified and should use the injected credentials
			assert.NotEqual(
				t,
				beforeR.Header.Get("Authorization"),
				afterR.Header.Get("Authorization"),
			)
			assert.Equal(
				t,
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
			assert.Equal(
				t,
				"https://s3.amazonaws.com",
				afterR.URL.String(),
			)
			// The Authorization should be modified and should use the injected credentials
			assert.Equal(
				t,
				// The expected Authorization header value has been manually calculated
				"AWS4-HMAC-SHA256 Credential=accessKeyIdValue/20210102/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=d75dd6809bc23c1045d7ea19f80bf18fc976b0943eed8aac5aa809810a6b2368",
				afterR.Header.Get("Authorization"),
			)

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
			assert.Contains(t, err.Error(), "AWS connection parameter")
			assert.Contains(t, err.Error(), "is not available")
		},
	},
}

func assertOnHeadersAfterSigning(t *testing.T, afterR *http.Request) {
	// X-Amz-Content-Sha256 is always recalculated
	assert.Equal(
		t,
		connectTestCaseBodySHA256,
		afterR.Header.Get("X-Amz-Content-Sha256"),
	)

	// Unsigned headers remain unchanged
	assert.Equal(
		t,
		"Unsigned-Header-1-Value",
		afterR.Header.Get("Unsigned-Header-1"),
	)
	assert.Equal(
		t,
		"Unsigned-Header-2-Value",
		afterR.Header.Get("Unsigned-Header-2"),
	)
}

func TestConnector_Connect(t *testing.T) {
	for _, testCase := range testCases {
		testCase.Run(t)
	}
}
