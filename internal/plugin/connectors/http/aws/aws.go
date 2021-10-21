package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// From https://github.com/aws/aws-sdk-go/blob/master/aws/signer/v4/v4.go#L77
const timeFormat = "20060102T150405Z"

// newAmzDate parses a date string using the AWS signer time format
func newAmzDate(amzDateStr string) (time.Time, error) {
	if amzDateStr == "" {
		return time.Time{}, fmt.Errorf("missing required header: %s", "X-Amz-Date")
	}

	return time.Parse(timeFormat, amzDateStr)
}

// requestMetadataFromAuthz parses an authorization header string and create a
// requestMetadata instance populated with the associated region, service
// name and signed headers
func requestMetadataFromAuthz(authorizationStr string) (*requestMetadata, error) {
	// Parse the following (line breaks added for readability):
	// AWS4-HMAC-SHA256 \
	// Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, \
	// SignedHeaders=host;range;x-amz-date, \
	// Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024
	//
	// See https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html

	// Validate form of entire authorization header
	tokens := strings.Split(authorizationStr, ", ")
	if len(tokens) != 3 || tokens[0] == "" || tokens[1] == "" || tokens[2] == "" {
		return nil, fmt.Errorf("malformed Authorization header")
	}

	// Extract region and service name from credential component
	credentialParts := strings.SplitN(tokens[0], "/", 5)
	if len(credentialParts) != 5 {
		return nil, fmt.Errorf("malformed credential component of Authorization header")
	}

	region := credentialParts[2]
	serviceName := credentialParts[3]

	// Extract signed headers from signed headers component
	signedHeaders := strings.Split(
		strings.TrimPrefix(tokens[1], "SignedHeaders="),
		";",
	)

	return &requestMetadata{
		region:        region,
		serviceName:   serviceName,
		signedHeaders: signedHeaders,
	}, nil
}

// requestMetadata captures the metadata of a signed AWS request: date, region, service
// name and signed headers
type requestMetadata struct {
	date          time.Time
	region        string
	serviceName   string
	signedHeaders []string
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

	// Parse date string
	//
	date, err := newAmzDate(amzDateStr)
	if err != nil {
		return nil, err
	}

	// Create request metadata by extracting service name and region from
	// Authorization header
	reqMeta, err := requestMetadataFromAuthz(authorizationStr)
	if err != nil {
		return nil, err
	}

	// Populate request metadata with date
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
	return err
}

// maybeSetAmzEndpoint, when the request URL is http://secretless.empty, sets the
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
// Note that if the client specifies an HTTP (not HTTPS, because Secretless does not proxy
// HTTPS requests) endpoint that is not http://secretless.empty it will be respected.
//
// Note: There is a plan to add a configuration option to instruct Secretless to
// upgrade the connect between Secretless and the target endpoint to TLS
func maybeSetAmzEndpoint(req *gohttp.Request, reqMeta *requestMetadata) error {
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
	req.Host = endpointURL.Hostname()

	return nil
}
