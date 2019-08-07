package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/signer/v4"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
)

// AWSHandler applies AWS signature authentication to the HTTP Authorization header.
type AWSHandler struct {
	plugin_v1.BaseHandler
}

// AWS4-HMAC-SHA256 Credential=AKIAJC5FABNOFVBKRWHA/20171103/us-east-1/ec2/aws4_request
var headerRegexp = regexp.MustCompile(`^AWS4-HMAC-SHA256 Credential=\w+\/\d+\/([\w-_]+)\/(\w+)\/aws4_request$`)

// From https://github.com/aws/aws-sdk-go/blob/master/aws/signer/v4/v4.go#L77
const (
	authHeaderPrefix = "AWS4-HMAC-SHA256"
	timeFormat       = "20060102T150405Z"
	shortTimeFormat  = "20060102"

	// emptyStringSHA256 is a SHA256 of an empty string
	emptyStringSHA256 = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
)

// Authenticate applies the "accessKeyId", "secretAccessKey" and optional "accessToken" credentials
// to the Authorization header, following the AWS signature format.
func (h AWSHandler) Authenticate(values map[string][]byte, r *http.Request) error {
	var err error
	var amzDate time.Time

	authorization := strings.Join(r.Header["Authorization"], "")
	amzDateStr := strings.Join(r.Header["X-Amz-Date"], "")

	// Don't sign the request when the original request is not signed.
	// Without an existing Authorization header, we can't determine required signing
	//   parameters such as the ServiceName.
	if authorization == "" {
		return nil
	}

	if amzDateStr == "" {
		return fmt.Errorf("Missing required header : X-Amz-Date")
	}
	if amzDate, err = time.Parse(timeFormat, amzDateStr); err != nil {
		return err
	}

	var accessKeyId, secretAccessKey, accessToken []byte
	var header string

	accessKeyId = values["accessKeyId"]
	if accessKeyId == nil {
		return fmt.Errorf("AWS connection parameter 'accessKeyId' is not available")
	}
	secretAccessKey = values["secretAccessKey"]
	if secretAccessKey == nil {
		return fmt.Errorf("AWS connection parameter 'secretAccessKey' is not available")
	}
	accessToken = values["accessToken"]

	creds := credentials.NewStaticCredentials(string(accessKeyId), string(secretAccessKey), string(accessToken))

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Parse the following (line breaks added):
	// AWS4-HMAC-SHA256 Credential=AKIAJC5FABNOFVBKRWHA/20171103/us-east-1/ec2/aws4_request, \
	//   SignedHeaders=content-type;host;x-amz-date, \
	//   Signature=c4a8ade09a5e0c644cc282311c36aae6c834596076ffde7db7d1195c7b454ed0
	if header, _, _, err = func(authorization string) (string, string, string, error) {
		tokens := strings.Split(authorization, ", ")
		if len(tokens) != 3 || tokens[0] == "" || tokens[1] == "" || tokens[2] == "" {
			return "", "", "", fmt.Errorf("Malformed Authorization header")
		}
		return tokens[0], tokens[1], tokens[2], nil
	}(authorization); err != nil {
		return err
	}

	matches := headerRegexp.FindStringSubmatch(header)
	if len(matches) != 3 {
		return fmt.Errorf("Malformed header section of Authorization header")
	}
	region := matches[1]
	serviceName := matches[2]

	signer := v4.NewSigner(creds)

	// TODO: Make this dependent on a build flag instead of handler flag
	// https://github.com/cyberark/secretless-broker/issues/593
	//
	// if h.GetConfig().Debug {
	// 	signer.Debug = aws.LogDebugWithSigning
	// 	signer.Logger = aws.NewDefaultLogger()
	// }

	if _, err := signer.Sign(r, bytes.NewReader(bodyBytes), serviceName, region, amzDate); err != nil {
		return err
	}

	// When the endpoint URL is http://secretless.empty
	// use default resolver to get service endpoint
	if r.URL.Scheme == "http" && r.URL.Host == "secretless.empty" {
		// The endpoint will use TLS if possible
		endpoint, _ := endpoints.DefaultResolver().EndpointFor(
			serviceName,
			region)

		endpointURL, _ := url.Parse(endpoint.URL)

		r.URL.Scheme = endpointURL.Scheme
		r.URL.Host = endpointURL.Host
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

	return nil
}

// AWSHandlerFactory instantiates a handler given HandlerOptions
func AWSHandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	return &AWSHandler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}
}
