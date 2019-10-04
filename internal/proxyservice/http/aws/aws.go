package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// From https://github.com/aws/aws-sdk-go/blob/master/aws/signer/v4/v4.go#L77
const timeFormat = "20060102T150405Z"

// AWS4-HMAC-SHA256 Credential=AKIAJC5FABNOFVBKRWHA/20171103/us-east-1/ec2/aws4_request
// See https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
var reForCredentialComponent = regexp.MustCompile(`^Credential=\w+\/\d+\/([\w-_]+)\/(\w+)\/aws4_request$`)

// newAmzDate parses a date string using the AWS signer time format
func newAmzDate(amzDateStr string) (time.Time, error) {
	if amzDateStr == "" {
		return time.Time{}, fmt.Errorf("missing required header: %s", "X-Amz-Date")
	}

	return time.Parse(timeFormat, amzDateStr)
}

// requestMetadataFromAuthz parses an authorization header string and create a
// requestMetadata instance populated with the associated region and service
// name
func requestMetadataFromAuthz(authorizationStr string) (*requestMetadata, error) {
	// extract credentials section of authorization header
	credentialsComponent, err := extractCredentialsComponent(authorizationStr)
	if err != nil {
		return nil, err
	}
	// validate credential component of authorization header, then extract region
	// and service name
	matches := reForCredentialComponent.FindStringSubmatch(credentialsComponent)
	if len(matches) != 3 {
		return nil, fmt.Errorf("malformed credential component of Authorization header")
	}

	return &requestMetadata{
		region:      matches[1],
		serviceName: matches[2],
	}, nil
}

// requestMetadata captures the metadata of a signed AWS request: date, region
// and service name
type requestMetadata struct {
	date        time.Time
	region      string
	serviceName string
}

// extractCredentialsComponent extracts the credentials component from an
// authorization header string
func extractCredentialsComponent(authorizationStr string) (string, error) {
	// Parse the following (line breaks added):
	// AWS4-HMAC-SHA256
	// Credential=AKIAJC5FABNOFVBKRWHA/20171103/us-east-1/ec2/aws4_request, \
	// SignedHeaders=content-type;host;x-amz-date, \
	// Signature=c4a8ade09a5e0c644cc282311c36aae6c834596076ffde7db7d1195c7b454ed0

	// validate form of entire authorization header
	tokens := strings.Split(authorizationStr, ", ")
	if len(tokens) != 3 || tokens[0] == "" || tokens[1] == "" || tokens[2] == "" {
		return "", fmt.Errorf("malformed Authorization header")
	}

	// trim prefix and return credential component
	return strings.TrimPrefix(tokens[0], "AWS4-HMAC-SHA256 "), nil
}

// newRequestMetadata parses the request headers to extract the metadata
// necessary to sign the request
func newRequestMetadata(r *gohttp.Request) (*requestMetadata, error) {
	authorizationStr := r.Header.Get("Authorization")
	amzDateStr := r.Header.Get("X-Amz-Date")

	// Without an existing Authorization header, we can't determine required
	// signing parameters such as the ServiceName.
	if authorizationStr == "" {
		return nil, nil
	}

	// parse date string
	//
	date, err := newAmzDate(amzDateStr)
	if err != nil {
		return nil, err
	}

	// create request metadata by extracting service name and region from
	// Authorization header
	reqMeta, err := requestMetadataFromAuthz(authorizationStr)
	if err != nil {
		return nil, err
	}

	// populate request metadata with date
	reqMeta.date = date

	return reqMeta, nil
}

// newAmzStaticCredentials generates static AWS credentials from a credentials
// map
func newAmzStaticCredentials(
	credentialsByID map[string][]byte,
) (*credentials.Credentials, error) {
	var accessKeyID, secretAccessKey, accessToken []byte

	accessKeyID = credentialsByID["accessKeyId"]
	if len(accessKeyID) == 0 {
		return nil, fmt.Errorf("AWS connection parameter 'accessKeyId' is not available")
	}

	secretAccessKey = credentialsByID["secretAccessKey"]
	if len(secretAccessKey) == 0 {
		return nil, fmt.Errorf("AWS connection parameter 'secretAccessKey' is not available")
	}

	accessToken = credentialsByID["accessToken"]

	return credentials.NewStaticCredentials(
		strings.TrimSpace(string(accessKeyID)),
		strings.TrimSpace(string(secretAccessKey)),
		strings.TrimSpace(string(accessToken)),
	), nil
}

// signRequest uses metadata and credentials to sign a request
func signRequest(
	r *gohttp.Request,
	reqMeta *requestMetadata,
	credentialsByID connector.CredentialValuesByID,
) error {
	// Create AWS static credentials using provided credentials
	amzCreds, err := newAmzStaticCredentials(credentialsByID)
	if err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	_, err = signer.NewSigner(amzCreds).Sign(
		r,
		bytes.NewReader(bodyBytes),
		reqMeta.serviceName,
		reqMeta.region,
		reqMeta.date,
	)
	if err != nil {
		return err
	}

	return nil
}

// setAmzEndpoint, when the request URL is http://secretless.empty, sets the
// request endpoint using the default AWS endpoint resolver. The resolver allows
// the connector to mimic a typical AWS client and provides a TLS endpoint where
// possible.
//
// An endpoint URL of http://secretless.empty signifies to use the default
// resolver to get the service endpoint. This measure is necessary to address
// the issue that clients usually speak to Amazon over TLS. However, this is an
// HTTP only proxy. In order to use this proxy a client has to use a dummy HTTP
// endpoint and then this connector can use the AWS SDK to resolve the endpoint
// in the same way the client might via a direct call to Amazon over HTTPS.
//
// Note that if the client to specifies an HTTP (not HTTPS) endpoint that is not
// http://secretless.empty it will be respected.
//
// Note: There is a plan to add a configuration option to instruct Secretless to
// upgrade the connect between Secretless and the target endpoint to TLS.
func setAmzEndpoint(req *gohttp.Request, reqMeta *requestMetadata) error {
	shouldSetEndpoint := req.URL.Scheme == "http" &&
		req.URL.Host == "secretless.empty"

	if !shouldSetEndpoint {
		return nil
	}

	endpoint, err := endpoints.DefaultResolver().EndpointFor(
		reqMeta.serviceName,
		reqMeta.region,
	)
	if err != nil {
		return err
	}

	endpointURL, err := url.Parse(endpoint.URL)
	if err != nil {
		return err
	}

	req.URL.Scheme = endpointURL.Scheme
	req.URL.Host = endpointURL.Host

	return nil
}
