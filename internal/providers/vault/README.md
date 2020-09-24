# Vault Provider

The Vault provider for Secretless can fetch secrets from configured secret
engines in [HashiCorp Vault](https://www.vaultproject.io). The provider is based
on the [Vault API client](https://pkg.go.dev/github.com/hashicorp/vault/api) in
Go. It reads the secret object from the configured path and returns the value
navigated to by the configured fields (or default field otherwise).

## Usage Documentation

The Vault provider is configured in the `secretless.yml` using:

```yaml
from: vault
get: /path/to/secret/in/vault
```

Or with explicit fields navigating to the value in the secret returned at path:

```yaml
from: vault
get: /path/to/secret/in/vault#navigate.to.this.field
```

The provider will read a secret (object) at a given path and returns the value
of field `value` (by default). By appending `#data.fieldName` to the path, the
provider will instead read the value at the field `fieldName` in the object
`data` in the secret (object) instead.

Below are some examples showing how to configure the provider for secrets.

### Example: API key from KV backends (v1 and v2)

Below is an excerpt of an example configuration for a fictional "Example
Service" that requires an API key, e.g. used in a request header. It gets the
API key from Vault's KV version 1 backend at path `kv/example-service` under the
secret's `value` field.

```yaml
version: 2
services:
  my_example_service:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      apikey:
        from: vault
        get: kv/example-service
        # gets path to API key in Vault, field 'value' holds the API key
    ...
```

A slightly different configuration explicitly sets the field `api-key` (instead
of the default `value`) to hold the API key.

```yaml
version: 2
services:
  my_example_service:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      apikey:
        from: vault
        get: kv/example-service#api-key
        # gets path to API key in Vault, field 'api-key' holds the API key
    ...
```

If the secret is stored in a KV v2 backend (mounted at `secret` by default), the
configuration must use the use the `data` segment in the path and the
`#data.api-key` suffix. This is behavior specific to KV v2 in Vault, see Vault
API docs.

```yaml
version: 2
services:
  my_example_service:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      apikey:
        from: vault
        get: secret/data/example-service#data.api-key
        # gets path to API key in Vault stored in the KV v2 secret engine
    ...
```

## Limitations

- Only token-based login to Vault supported at the moment.
- Only secrets that are "read" in Vault are supported at the moment. Backends
  that require "writes" to obtain the secret (e.g. PKI, dynamic database
  credentials) are not supported at the moment.
- Backends that have multiple values change simultaneously (e.g. client id and
  secret, database username and password) are not supported at the moment.
- Limited support for KV v2 secret engine, only latest version of a secret can
  be retrieved.
