package aws

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"io"
	gohttp "net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sts "github.com/aws/aws-sdk-go-v2/service/sts"
	smendpoints "github.com/aws/smithy-go/endpoints"
)

// From https://github.com/aws/aws-sdk-go-v2/blob/c28a6f44b3a2dfa2495cea65beedfa7958f6dda6/aws/signer/internal/v4/const.go#L30
const timeFormat = "20060102T150405Z"

// newAmzDate parses a date string using the AWS signer time format
func newAmzDate(amzDateStr string) (time.Time, error) {
	if amzDateStr == "" {
		return time.Time{}, fmt.Errorf("missing required header: %s", "X-Amz-Date")
	}
	tm, err := time.Parse(timeFormat, amzDateStr)
	if tm.IsZero() {
		tm = time.Now()
	}
	return tm, err
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

// signRequest uses metadata and credentials to sign a request (older signer/v4 API shape).

func signRequest(
	req *gohttp.Request,
	reqMeta *requestMetadata,
	credentialsByID connector.CredentialValuesByID,
) error {
	accessKeyID := strings.TrimSpace(string(credentialsByID["accessKeyId"]))
	secretAccessKey := strings.TrimSpace(string(credentialsByID["secretAccessKey"]))
	sessionToken := strings.TrimSpace(string(credentialsByID["accessToken"])) // optional
	if accessKeyID == "" || secretAccessKey == "" {
		return fmt.Errorf("missing required AWS credentials: accessKeyId/secretAccessKey")
	}
	creds := aws.Credentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		Source:          "secretless-aws-connector",
	}

	payloadHash := strings.TrimSpace(req.Header.Get("X-Amz-Content-Sha256"))
	if payloadHash == "" || strings.EqualFold(payloadHash, "this-will-be-generated-at-signing") {
		var bodyBytes []byte
		if req.Body != nil {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, req.Body); err != nil {
				return fmt.Errorf("reading request body for hashing: %w", err)
			}
			bodyBytes = buf.Bytes()
			// restore body for downstream usage
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			// keep ContentLength aligned; will be neutralized if we exclude it
			req.ContentLength = int64(len(bodyBytes))
		} else {
			// empty payload
			bodyBytes = nil
		}
		sum := sha256.Sum256(bodyBytes)
		payloadHash = fmt.Sprintf("%x", sum[:])
		req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	}

	containsFold := func(list []string, want string) bool {
		for _, s := range list {
			if strings.EqualFold(strings.TrimSpace(s), strings.TrimSpace(want)) {
				return true
			}
		}
		return false
	}
	removeContentLength := !containsFold(reqMeta.signedHeaders, "content-length")

	var originalContentLength string
	var hadContentLength bool
	var originalContentLengthValue int64
	if removeContentLength {
		originalContentLength = req.Header.Get("Content-Length")
		hadContentLength = originalContentLength != ""
		if hadContentLength {
			if cl, err := strconv.ParseInt(originalContentLength, 10, 64); err == nil {
				originalContentLengthValue = cl
			}
		}
		req.Header.Del("Content-Length")
		req.ContentLength = 0
		req.TransferEncoding = nil
	}

	signer := v4.NewSigner()
	if err := signer.SignHTTP(
		context.Background(),
		creds,
		req,
		payloadHash,
		reqMeta.serviceName,
		reqMeta.region,
		reqMeta.date.UTC(), // explicit signing time
	); err != nil {
		// restore on error
		if removeContentLength && hadContentLength {
			req.Header.Set("Content-Length", originalContentLength)
			req.ContentLength = originalContentLengthValue
		}
		return err
	}

	// Restore header/fields (optional; transport will set them anyway)
	if removeContentLength && hadContentLength {
		req.Header.Set("Content-Length", originalContentLength)
		req.ContentLength = originalContentLengthValue
	}
	return nil
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
	shouldSetEndpoint := req.URL.Scheme == "http" && req.URL.Host == "secretless.empty"
	if !shouldSetEndpoint {
		return nil
	}

	// v2 endpoint resolution via config
	ctx := context.Background()

	// Resolve endpoint via Endpoints v2 for the target service+region.
	ep, err := resolveEndpointV2(ctx, reqMeta.serviceName, reqMeta.region)
	if err != nil {
		return err
	}

	// smithy-go endpoints.Endpoint carries a full URI
	req.URL.Scheme = ep.URI.Scheme
	req.URL.Host = ep.URI.Host
	req.Host = ep.URI.Hostname()
	return nil
}

// Resolve via the service's v2 resolver. For services whose module version
// is pre-v2 (no EndpointParameters/NewDefaultEndpointResolverV2), we fall back
// to env-provided base endpoint.
func resolveEndpointV2(ctx context.Context, serviceID, region string) (smendpoints.Endpoint, error) {
	useDualStack := isTrueEnv("AWS_USE_DUALSTACK_ENDPOINT")
	useFIPS := isTrueEnv("AWS_USE_FIPS_ENDPOINT")

	// Prefer service-specific endpoint if present; otherwise use the global.
	base := baseEndpointFor(serviceID)
	if base == "" {
		base = os.Getenv("AWS_ENDPOINT_URL")
	}

	switch serviceID {
	case s3.ServiceID:
		// Older service modules expect *string/*bool EndpointParameters.
		// If your module expects value types, use the alt version shown below.
		params := s3.EndpointParameters{
			Region:       aws.String(region),
			UseDualStack: aws.Bool(useDualStack),
			UseFIPS:      aws.Bool(useFIPS),
			Endpoint:     aws.String(base),
		}
		return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)

	case sts.ServiceID:
		params := sts.EndpointParameters{
			Region:       aws.String(region),
			UseDualStack: aws.Bool(useDualStack),
			UseFIPS:      aws.Bool(useFIPS),
			Endpoint:     aws.String(base),
		}
		return sts.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)

	case secretsmanager.ServiceID:
		// If your secretsmanager module is pre-v2 (no EndpointParameters),
		// comment this block out and let the fallback below return an env-only endpoint.
		params := secretsmanager.EndpointParameters{
			Region:       aws.String(region),
			UseDualStack: aws.Bool(useDualStack),
			UseFIPS:      aws.Bool(useFIPS),
			Endpoint:     aws.String(base),
		}
		return secretsmanager.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
	}

	// Fallback: construct endpoint strictly from env (no per-service resolver).
	if base == "" {
		return smendpoints.Endpoint{}, fmt.Errorf("no resolver for serviceID %q and no AWS_ENDPOINT_URL_%s provided", serviceID, strings.ToUpper(serviceID))
	}
	u, err := url.Parse(base)
	if err != nil {
		return smendpoints.Endpoint{}, fmt.Errorf("invalid endpoint URL %q: %w", base, err)
	}
	return smendpoints.Endpoint{URI: *u}, nil
}

func isTrueEnv(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	return strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
}

// AWS_ENDPOINT_URL_<SERVICE> or empty string
func baseEndpointFor(serviceID string) string {
	if serviceID == "" {
		return ""
	}
	key := "AWS_ENDPOINT_URL_" + strings.ToUpper(serviceID)
	return strings.TrimSpace(os.Getenv(key))
}
