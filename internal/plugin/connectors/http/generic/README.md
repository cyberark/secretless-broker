# Generic HTTP Authentication

Secretless comes with built-in HTTP connectors that let you to connect to
services like AWS, or any service that uses Basic authentication.  But what if
you want to connect to a service that Secretless has no connector for?

The Generic HTTP connector can help you.  

It can inject any credential you want directly into any HTTP header.  You
define the header name and the credential you want injected.  You can even
concatenate credentials together, or encode them in base64.

## End User Documentation

Using the Generic HTTP connector is just a matter of configuring your
`secretless.yml`.

It's easiest to learn how this works by example.

### Example Configurations

We'll start with two example configurations, and then explain each of the
relevant details.

#### Header with single credential

Here's an simple example configuration for a fictional "Example Service" that
requires an single API key credential in a header called `X-ApiKey`.

```yaml
version: 2
services:
  my_example_service:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      apikey:
        from: conjur
        get: my-services-api-key  # the id of your API key within conjur
    config:
      credentialValidations:
        apikey: '^[A-Z0-9]+$'     # valid API keys consist of uppercase letters
      headers:                    # and digits only
        "X-ApiKey": "{{ .apikey }}"
      forceSSL: true
      authenticateURLsMatching:
        - ^http                   # apply this connector to all requests
```

#### Reimplementing Basic Authentication

Of course, if you needed Basic Authentication in a real application, you'd use
the built-in `basic_auth` connector.  But the Generic HTTP connector is
powerful enough to provide the same functionality, and it's instructive to see
how it _could_ be done:

```yaml
version: 2
services:
  service_requiring_basic_auth:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      username:
        from: conjur
        get: someuser
      password:
        from: conjur
        get: somepassword
      address:
        from: conjur
        get: address
    config:
      credentialValidations:
        username: '[^:]+'    # username cannot contain a colon
      headers:
        Authorization: "Basic {{ printf \"%s:%s\" .username .password | base64 }}"
      queryParams:
        location: "{{ .address }}"
      forceSSL: true
      authenticateURLsMatching:
        - ^http
```

### How It Works

The two important features to note are the `headers` and
`credentialValidations` sections of the `config`:

#### `headers`

This is the key section.

The _names_ of the headers are defined by the yaml keys.  In the examples
above, these header names are `X-ApiKey` and `Authorization`, respectively.

The header _values_ are defined using a [Go text
template](https://golang.org/pkg/text/template/), as defined in the
`text/template` package.

You can refer to your credentials in this template using the credential name
preceded by a `.` (eg, `.username` and `.password` refer to the credentials
`username` and `password`). At runtime, Secretless will replace these
credential references with your real credentials.

As you can see in the Basic auth example, the `text/template` package has
powerful transformation features.  You can use `printf` for formatting and
compose functions using pipes `|`.  See the text template package docs linked
above for detailed information on these and other features.

### `queryParams`

Like `headers`, this is another key section. The `queryParams` section
is used to generate a query string, which is appended to your existing URL
without replacing any existing query parameters.

The _keys_ of the queryParams are defined by the yaml keys.  In the examples
above, the query parameter key is `location`.

The query parameter _values_ are defined using a [Go text
template](https://golang.org/pkg/text/template/), as defined in the
`text/template` package.

In the above example, let us say that your request URL looks like the following,

```http
http://anything.com/foo?fruit=apple
```

After proxying through secretless, your request URL would look like the following,

```http
http://anything.com/foo?location=valueofaddress&fruit=apple
```

You can refer to your credentials in this template using the credential name
preceded by a `.` (eg, `.address` will refer to the credential
`address`). At runtime, Secretless will replace these
credential references with your real credentials.

As you can see in the Basic auth example, the `text/template` package has
powerful transformation features.  You can use `printf` for formatting and
you can compose functions using pipes `|`.  See the text template package docs linked
above for detailed information on these and other features.

### `oauth1`

Like `headers` this is another key section. The `oauth1` section is used to generate
an OAuth1 `Authorization` header.

**Note: Declaring an `Authorization` header in the `config`
 while `oauth1` is present will throw an error:
 `authorization header already exists, cannot override header`**

There are four required _keys_ for the `oauth1` section which are as follows:

1. `consumer_key` - A value used by the consumer to identify itself to
   the service provider.
1. `consumer_secret` - A secret used by the consumer to establish ownership
   of the consumer key.
1. `token` - A value used by the consumer to gain access to the protected
   resources on behalf of the user, instead of using the user’s service
   provider credentials.
1. `token_secret` - A secret used by the consumer to establish ownership of
   a given token.

The oauth1 _values_ are defined using a [Go text
template](https://golang.org/pkg/text/template/), as defined in the
`text/template` package.

You can refer to your credentials in this template using the credential name
preceded by a `.` (e.g. `.consumer_key` and `.token` refer to the credentials
`consumer_key` and `token`). At runtime, Secretless will replace these
credential references with your real credentials.

For instance:

```yaml

version: 2
services:
  oauth1-service:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      consumer_key:
        from: conjur
        get: somekey
      consumer_secret:
        from: conjur
        get: someconsumersecret
      token:
        from: conjur
        get: sometoken
      token_secret:
        from: conjur
        get: sometokensecret
    config:
      oauth1:
        consumer_key: "{{ .consumer_key }}"
        consumer_secret: "{{ .consumer_secret }}"
        token: "{{ .token }}"
        token_secret: "{{ .token_secret }}"
      forceSSL: true
      authenticateURLsMatching:
        - ^http
```

#### `credentialValidations`

This section lets you use regular expressions to define validations for the
header values.  

For example, in the first example above, the line:

```yaml
apikey: '^[A-Z0-9]+$'
```

tells us that we expect the `apikey` to consist solely of uppercase letters and
digits.  If this rule is violated at runtime, Secretless will log an
appropriate error.

### Limitations

Some HTTP APIs require more involved authentication.  For example, they might
require you to read the HTTP body content and create a hashed signature of it
using a secret key.

For now, the Generic HTTP connector does not support use cases like these,
though it may in the future.

Currently, it can do the following:

- Create a header with any name
- Populate that header with any credential
- Populate that header any combination of credentials and literal strings
- Populate that header any supported transformation of any combination of
  credentials and fixed strings

Basically, you are limited only by what's possible in the Go `text/template`
package, and by the additional functions Secretless makes available.

Currently, the only additonal function beyond the defaults of `text/template`
is `base64`, but we plan to add more and can do so easily.  If there's one
you'd like to see, please create an issue or pull request.

## Developer Documentation

For developers, the `generic` package makes writing new HTTP connectors easy.
You can create a new connector with just a bit of boilerplate and a single
declarative struct to define the Authorization headers.

### Defining your connector

You'll define your generic connector using the `ConfigYAML` type.  Essentially,
`ConfigYAML` lets you use Go code to write the same configuration that
end-users write in `secretless.yml`.  

Let's take a look at the Basic auth connector to see how this works:

```go
// Taken from:
// internal/plugin/connectors/http/basicauth/plugin.go
//
// Also note:
// "NewConnectorConstructor" and "ConfigYAML" are defined in:
// /internal/plugin/connectors/http/generic/external_api.go

newConnector, err := generichttp.NewConnectorConstructor(
  &generichttp.ConfigYAML{
    CredentialValidations: map[string]string{
      "username": "[^:]+",
    },
    Headers: map[string]string{
      "Authorization": "Basic {{ printf \"%s:%s\" .username .password | base64 }}",
    },
  },
)
```

Note how this code exactly mirrors the yaml from the end-user example.  Please
refer to those docs above for an explanation of the `CredentialValidations` and
`Headers` formats.

`ConfigYAML`, then, defines your connector, and `NewConnectorConstructor` is a
convenience function to transform that definition into an `http.Plugin`.

### Adding your connector

While the heart of creating a connector is the `ConfigYAML` definition
described above, it also requires a few pieces of boilerplate.

Refer to the Basic auth connector
(`internal/plugin/connectors/http/basicauth/plugin.go`) as an example when
working through the steps below.

Here are the steps, in detail:

1. Create a new directory for your connector under `/internal/plugin/connectors/http`.
2. Create a `plugin.go` file with an appropriate package name.
3. Write your `PluginInfo()` and `GetHTTPPlugin()` functions.
4. Create a new directory under `/test/connector/http` for your integration tests.
5. Write the integration tests themselves.  Please refer to [our docs for adding new integration tests](
   https://github.com/cyberark/secretless-broker/blob/master/CONTRIBUTING.md#adding-new-integration-tests)
