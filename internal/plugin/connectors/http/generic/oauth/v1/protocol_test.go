package oauth1protocol

import (
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractOAuthParams(t *testing.T) {
	paramMap := map[string]string{
		consumerKey:          "consumerKey",
		consumerSecret:       "consumerSecret",
		token:                "token",
		tokenSecret:          "tokenSecret",
		oauthNonce:           "oauth_nonce",
		oauthSignatureMethod: "oauth_signature_method",
		oauthVersion:         "oauth_version",
		oauthTimestamp:       "oauth_timestamp",
	}
	result := extractOAuthParams(paramMap)
	// Nonce and TimeStamp are generated at runtime and unknown
	result.TimeStamp = ""
	result.Nonce = ""

	want := OAuth1{
		ConsumerKey:     "consumerKey",
		ConsumerSecret:  "consumerSecret",
		Nonce:           "",
		SignatureMethod: "HMAC-SHA1",
		TimeStamp:       "",
		Token:           "token",
		TokenSecret:     "tokenSecret",
		Version:         "1.0",
	}
	assert.Equal(t, want, result)
}

func createRequestBody(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func Test_collectParameters(t *testing.T) {
	type args struct {
		oauth1 OAuth1
		r      *gohttp.Request
	}

	params := args{
		oauth1: OAuth1{
			ConsumerKey:     "xvz1evFS4wEEPTGEFPHBog",
			Nonce:           "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg",
			SignatureMethod: "HMAC-SHA1",
			TimeStamp:       "1318622958",
			Token:           "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
			Version:         "1.0",
		},
		r: &gohttp.Request{
			Method: "POST",
			URL: &url.URL{
				Scheme:   "http",
				Host:     "example.com",
				Path:     "/wp-json/wp/v2/posts",
				RawQuery: "include_entities=true",
			},
			Header: map[string][]string{
				"Authorization": {"doesn't matter"},
				"Content-Type":  {"application/x-www-form-urlencoded"},
			},
			Body: createRequestBody("status=Hello%20Foo%20%2B%20Bar%2C%20a%20signed%20OAuth%20request%21"),
		},
	}

	result := collectParameters(params.oauth1, params.r)

	want := map[string][]string{
		"include_entities":       {"true"},
		"oauth_consumer_key":     {"xvz1evFS4wEEPTGEFPHBog"},
		"oauth_nonce":            {"kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg"},
		"oauth_signature_method": {"HMAC-SHA1"},
		"oauth_timestamp":        {"1318622958"},
		"oauth_token":            {"370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb"},
		"oauth_version":          {"1.0"},
		"status":                 {"Hello%20Foo%20%2B%20Bar%2C%20a%20signed%20OAuth%20request%21"},
	}

	assert.Equal(t, want, result)
}

func Test_generateParameterString(t *testing.T) {
	type args struct {
		paramMap map[string][]string
	}
	tests := []struct {
		description string
		args        args
		want        string
		wantErr     bool
	}{
		{
			description: "keys with single values",
			args: args{
				paramMap: map[string][]string{
					"oauth_consumer_key":     {"key"},
					"oauth_nonce":            {"nonce"},
					"oauth_signature_method": {"HMAC-SHA1"},
					"oauth_timestamp":        {"123456789"},
					"oauth_token":            {"token"},
					"oauth_version":          {"1.0"},
				},
			},
			want: "oauth_consumer_key=key&oauth_nonce=nonce&oauth_signature_method=HMAC-SHA1&oauth_timestamp=123456789&oauth_token=token&oauth_version=1.0",
		},
		{
			description: "key with multiple values",
			args: args{
				paramMap: map[string][]string{
					"test_multi_var[]": {"multi1", "multi2"},
					"test_single_var":  {"single"},
				},
			},
			want: "test_multi_var[]=multi1&test_multi_var[]=multi2&test_single_var=single",
		},
		{
			description: "sorts map values",
			args: args{
				paramMap: map[string][]string{
					"c": {"3"},
					"a": {"1"},
					"b": {"2"},
				},
			},
			want: "a=1&b=2&c=3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := generateParameterString(tt.args.paramMap)
			assert.Equal(t, tt.want, result)
		},
		)
	}
}

func Test_generateBaseString(t *testing.T) {
	type args struct {
		paramString string
		r           *gohttp.Request
	}
	tests := []struct {
		description string
		args        args
		want        string
	}{
		{
			description: "all components supplied",
			args: args{
				paramString: "include_entities=true&oauth_consumer_key=xvz1evFS4wEEPTGEFPHBog&oauth_nonce=kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg&oauth_signature_method=HMAC-SHA1&oauth_timestamp=1318622958&oauth_token=370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb&oauth_version=1.0&status=Hello%20Foo%20%2B%20Bar%2C%20a%20signed%20OAuth%20request%21",
				r: &gohttp.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme: "https",
						Host:   "api.twitter.com",
						Path:   "/1.1/statuses/update.json",
					},
				},
			},
			want: "POST&https%3A%2F%2Fapi.twitter.com%2F1.1%2Fstatuses%2Fupdate.json&include_entities%3Dtrue%26oauth_consumer_key%3Dxvz1evFS4wEEPTGEFPHBog%26oauth_nonce%3DkYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1318622958%26oauth_token%3D370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb%26oauth_version%3D1.0%26status%3DHello%2520Foo%2520%252B%2520Bar%252C%2520a%2520signed%2520OAuth%2520request%2521",
		},
		{
			description: "no components supplied",
			args: args{
				paramString: "include_entities=true&oauth_consumer_key=xvz1evFS4wEEPTGEFPHBog&oauth_nonce=kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg&oauth_signature_method=HMAC-SHA1&oauth_timestamp=1318622958&oauth_token=370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb&oauth_version=1.0&status=Hello%20Foo%20%2B%20Bar%2C%20a%20signed%20OAuth%20request%21",
				r: &gohttp.Request{
					Method: "",
					URL: &url.URL{
						Scheme: "",
						Host:   "",
						Path:   "",
					},
				},
			},
			want: "&%3A%2F%2F&include_entities%3Dtrue%26oauth_consumer_key%3Dxvz1evFS4wEEPTGEFPHBog%26oauth_nonce%3DkYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1318622958%26oauth_token%3D370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb%26oauth_version%3D1.0%26status%3DHello%2520Foo%2520%252B%2520Bar%252C%2520a%2520signed%2520OAuth%2520request%2521",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := generateBaseString(tt.args.r, tt.args.paramString)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_generateNonce(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		description string
		args        args
		want        int
	}{
		{
			description: "Generates a 16 length nonce",
			args: args{
				length: 16,
			},
			want: 16,
		},
		{
			description: "Generates a 32 length nonce",
			args: args{
				length: 32,
			},
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := len(generateNonce(tt.args.length, nonceCharset))
			assert.True(t, result == tt.args.length)
		})
	}
}

func Test_generateSigningKey(t *testing.T) {
	type args struct {
		oauth1 OAuth1
	}
	tests := []struct {
		description string
		args        args
		want        string
		wantErr     bool
	}{
		{
			description: "both keys",
			args: args{
				OAuth1{
					ConsumerSecret: "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
					TokenSecret:    "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE",
				},
			},
			want: "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE",
		},
		{
			description: "no token secret",
			args: args{
				OAuth1{
					ConsumerSecret: "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
					TokenSecret:    "",
				},
			},
			want: "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&",
		},
		{
			description: "both empty",
			args: args{
				OAuth1{
					ConsumerSecret: "",
					TokenSecret:    "",
				},
			},
			want: "&",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := generateSigningKey(tt.args.oauth1)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_constructOAuthString(t *testing.T) {
	t.Run("creates expected header", func(t *testing.T) {
		oauth1 := OAuth1{
			ConsumerKey:     "xvz1evFS4wEEPTGEFPHBog",
			ConsumerSecret:  "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
			Nonce:           "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg",
			Signature:       "tnnArxj06cWHq44gCs1OSKk%2FjLY%3D",
			SignatureMethod: "HMAC-SHA1",
			TimeStamp:       "1318622958",
			Token:           "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
			TokenSecret:     "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE",
			Version:         "1.0",
		}

		result, err := constructOAuthString(oauth1)

		want := "OAuth oauth_consumer_key=\"xvz1evFS4wEEPTGEFPHBog\", oauth_nonce=\"kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg\", oauth_signature=\"tnnArxj06cWHq44gCs1OSKk%2FjLY%3D\", oauth_signature_method=\"HMAC-SHA1\", oauth_timestamp=\"1318622958\", oauth_token=\"370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb\", oauth_version=\"1.0\""

		assert.NoError(t, err)
		assert.Equal(t, want, result)
	})
}

func Test_encodeURI(t *testing.T) {
	type args struct {
		stringToEncode string
	}
	tests := []struct {
		description string
		args        args
		want        string
	}{
		{
			description: "snowman emoji",
			args: args{
				stringToEncode: "☃",
			},
			want: "%E2%98%83",
		},
		{
			description: "ALPHA",
			args: args{
				stringToEncode: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			},
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			description: "DIGIT",
			args: args{
				stringToEncode: "1234567890",
			},
			want: "1234567890",
		},
		{
			description: "special characters not to encode",
			args: args{
				stringToEncode: "-._~",
			},
			want: "-._~",
		},
		{
			description: "special characters to encode",
			args: args{
				stringToEncode: "`!@#$%^&*()+=[{]}\\|;:'\",<>/? ",
			},
			want: "%60%21%40%23%24%25%5E%26%2A%28%29%2B%3D%5B%7B%5D%7D%5C%7C%3B%3A%27%22%2C%3C%3E%2F%3F%20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := encodeURI(tt.args.stringToEncode)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_checkRequiredOAuthParams(t *testing.T) {
	type args struct {
		paramMap map[string]string
	}
	tests := []struct {
		description string
		args        args
		want        error
		wantErr     bool
	}{
		{
			description: "all values present: no error",
			args: args{
				paramMap: map[string]string{
					"consumer_key":    "conKey",
					"consumer_secret": "conSecret",
					"token":           "apiToken",
					"token_secret":    "tokSecret",
				},
			},
			want: nil,
		},
		{
			description: "missing value: consumer_key",
			args: args{
				paramMap: map[string]string{
					"consumer_secret": "conSecret",
					"token":           "apiToken",
					"token_secret":    "tokSecret",
				},
			},
			want: fmt.Errorf("required oAuth1 parameter 'consumer_key' not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := checkRequiredOAuthParams(tt.args.paramMap)
			assert.Equal(t, tt.want, result)
		},
		)
	}
}
