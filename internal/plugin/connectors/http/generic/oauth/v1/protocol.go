package oauth1protocol

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	gohttp "net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// OAuth1 is a struct of all OAuth1 specific parameters
type OAuth1 struct {
	ConsumerKey     string
	ConsumerSecret  string
	Nonce           string
	Signature       string
	SignatureMethod string
	TimeStamp       string
	Token           string
	TokenSecret     string
	Version         string
}

// OAuth Header Key Names
const (
	oauthConsumerKey     = "oauth_consumer_key"
	oauthNonce           = "oauth_nonce"
	oauthSignatureMethod = "oauth_signature_method"
	oauthTimestamp       = "oauth_timestamp"
	oauthToken           = "oauth_token"
	oauthVersion         = "oauth_version"
)

// OAuth Header Variables
const (
	// Nonce doesn't have a specified charset or length
	// Reference: https://tools.ietf.org/html/rfc5849#section-3.3
	nonceCharset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	nonceLength = 16
	// Currently, this feature supports HMAC-SHA1, but there is an issue logged to
	// support more methods.
	// Reference: https://github.com/cyberark/secretless-broker/issues/1324
	oauthSignatureMethodType = "HMAC-SHA1"
	oauthVersionNumber       = "1.0"
)

// Required Config YAML Variables
const (
	consumerKey    = "consumer_key"
	consumerSecret = "consumer_secret"
	token          = "token"
	tokenSecret    = "token_secret"
)

var requiredConfigParams = []string{
	consumerKey,
	consumerSecret,
	token,
	tokenSecret,
}

func generateNonce(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	randomChars := make([]byte, length)
	for index := range randomChars {
		randomChars[index] = charset[seededRand.Intn(len(charset))]
	}
	return string(randomChars)
}

// checkRequiredOAuthParams returns an error if a key from
// "requiredConfigParams" was not included in the supplied config
// returns the first missing key found
func checkRequiredOAuthParams(params map[string]string) error {
	for _, requiredParam := range requiredConfigParams {
		if len(params[requiredParam]) < 1 {
			return fmt.Errorf("required oAuth1 parameter '%s' not found", requiredParam)
		}
	}
	return nil
}

// extractOAuthParams extracts the required oAuth1 parameters from
// the config and assigns the other required paramters
func extractOAuthParams(params map[string]string) OAuth1 {
	oauth := OAuth1{
		ConsumerKey:     params[consumerKey],
		ConsumerSecret:  params[consumerSecret],
		Nonce:           generateNonce(nonceLength, nonceCharset),
		SignatureMethod: oauthSignatureMethodType,
		Token:           params[token],
		TokenSecret:     params[tokenSecret],
		TimeStamp:       strconv.Itoa(int(time.Now().Unix())),
		Version:         oauthVersionNumber,
	}

	return oauth
}

// encodeURI percent encodes a string and changes encoding of <SPACE> from "+" to "%20"
// Reference: https://tools.ietf.org/html/rfc5849#section-3.6
func encodeURI(stringToEncode string) string {
	return strings.ReplaceAll(url.QueryEscape(stringToEncode), "+", "%20")
}

// collectParameters creates a map of body and query parameters by percent encoding all
// keys and values and sorting values post-encoding
func collectParameters(oauth1 OAuth1, r *gohttp.Request) map[string][]string {
	// Create two copies of the request body
	// Since r.Body is io.ReadCloser the body is empty after reading.
	// Need to replace r.Body after r.ParseForm() dumps the body contents
	buf, _ := ioutil.ReadAll(r.Body)
	originalBody := ioutil.NopCloser(bytes.NewBuffer(buf))
	copyBody := ioutil.NopCloser(bytes.NewBuffer(buf))
	r.Body = originalBody

	paramMap := make(map[string][]string, 0)
	// get query and body params and URL encode keys and
	// values sort values after encoding
	r.ParseForm()
	for paramKey, paramValues := range r.Form {
		encodedKey := encodeURI(paramKey)
		paramMap[encodedKey] = paramValues

		// Keys with multiple values require the values
		// to be sorted post-percent encoding
		for valueIndex, value := range paramMap[encodedKey] {
			encodedValue := encodeURI(value)
			paramMap[encodedKey][valueIndex] = encodedValue
		}
		sort.Strings(paramMap[encodedKey])
	}
	r.Body = copyBody

	// these keys are already URL safe and don't need sorted(1 value)
	paramMap[oauthConsumerKey] = []string{encodeURI(oauth1.ConsumerKey)}
	paramMap[oauthNonce] = []string{encodeURI(oauth1.Nonce)}
	paramMap[oauthSignatureMethod] = []string{encodeURI(oauth1.SignatureMethod)}
	paramMap[oauthTimestamp] = []string{encodeURI(oauth1.TimeStamp)}
	paramMap[oauthToken] = []string{encodeURI(oauth1.Token)}
	paramMap[oauthVersion] = []string{encodeURI(oauth1.Version)}

	return paramMap
}

// generateParameterString creates a Parameter String from
// sorting an encoded map of parameters and values
// Reference: https://tools.ietf.org/html/rfc5849#section-3.5
func generateParameterString(paramMap map[string][]string) string {
	var parameterString strings.Builder

	// make map of keys and sort to append lexicographically later
	keys := make([]string, 0, len(paramMap))
	for key := range paramMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// create string
	for keyIndex, key := range keys {
		// Write keys and value pairs to `parameterString`
		// ex: key0=value&key1=value&key2=value
		for valueIndex, value := range paramMap[key] {
			// Do not add '&' on first value
			if keyIndex > 0 || valueIndex > 0 {
				parameterString.WriteString("&")
			}
			parameterString.WriteString(key + "=" + value)
		}
	}

	return parameterString.String()
}

// generateBaseString combines parameters to create the
// pre-hash base string for the signature
// Reference: https://tools.ietf.org/html/rfc5849#section-3.4.1
func generateBaseString(r *gohttp.Request, paramString string) string {
	method := strings.ToUpper(r.Method)
	urlString := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)
	return fmt.Sprintf("%s&%s&%s", method, encodeURI(urlString), encodeURI(paramString))
}

// generateSigningKey concats the consumer secret and
// token secret with an ampersand, and encodes them
// Reference: https://tools.ietf.org/html/rfc5849#section-3.4
func generateSigningKey(oauth1 OAuth1) string {
	return fmt.Sprintf("%s&%s", encodeURI(oauth1.ConsumerSecret), encodeURI(oauth1.TokenSecret))
}

// hashSignature creates a HMAC-SHA1 hash
// Reference: https://tools.ietf.org/html/rfc5849#section-3.4.2
func hashSignature(signingKey string, baseString string) string {
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseString))
	return encodeURI(base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

// constructOAuthString creates the Authorization header
// that is needed for the final request
// Reference: https://tools.ietf.org/html/rfc5849#section-3.5.1
func constructOAuthString(oauth1 OAuth1) (string, error) {
	tmpl, err := template.New("oauth").Parse("OAuth " +
		"oauth_consumer_key=\"{{.ConsumerKey}}\", " +
		"oauth_nonce=\"{{.Nonce}}\", " +
		"oauth_signature=\"{{.Signature}}\", " +
		"oauth_signature_method=\"{{.SignatureMethod}}\", " +
		"oauth_timestamp=\"{{.TimeStamp}}\", " +
		"oauth_token=\"{{.Token}}\", " +
		"oauth_version=\"{{.Version}}\"")
	if err != nil {
		return "", fmt.Errorf("could not create the OAuth template: %s", err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, oauth1)
	if err != nil {
		return "", fmt.Errorf("could not turn the OAuth template into a string: %s", err)
	}

	return tpl.String(), nil
}

// CreateOAuth1Header calls a series of methods to construct
// the oAuth1 'Authorization' Header String
func CreateOAuth1Header(params map[string]string, r *gohttp.Request) (string, error) {

	// check for required params
	err := checkRequiredOAuthParams(params)
	if err != nil {
		return "", err
	}

	oauth1 := extractOAuthParams(params)

	parameters := collectParameters(oauth1, r)
	paramString := generateParameterString(parameters)
	baseString := generateBaseString(r, paramString)
	signingKey := generateSigningKey(oauth1)

	oauth1.Signature = hashSignature(signingKey, baseString)

	oauthString, err := constructOAuthString(oauth1)
	if err != nil {
		return "", err
	}

	return oauthString, nil
}
