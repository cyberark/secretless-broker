---
title: Handlers
id: handlers
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/overview.html
---

## Overview

When the Secretless Broker receives a new request on a defined Listener, it automatically passes the request on to the Handler defined in the Secretless Broker configuration for processing. Each Listener in the Secretless Broker configuration should therefore have a corresponding Handler.

The Handler configuration specifies the Listener that the Handler is handling connections for and any credentials that will be needed for that connection. Several credential sources are currently supported; see the [Credential Providers](/docs/reference/providers/overview.html) section for more information.

The example below defines a Handler to process connection requests from the `pg_socket` Listener, and it has three credentials: `address`, `username`, and `password`. The `address` and `username` are literally specified in this case, and the `password` is taken from the environment of the running Secretless Broker process.
```yaml
handlers:
  - name: pg_via_socket
    listener: pg_socket
    credentials:
      - name: address
        provider: literal
        id: pg:5432
      - name: username
        provider: literal
        id: myuser
      - name: password
        provider: env
        id: PG_PASSWORD
```

In production you would want your credential information to be pulled from a vault, and the Secretless Broker currently supports multiple vault Credential Providers.

When a Handler receives a new connection requests, it retrieves any required credentials using the specified Provider(s), injects the correct authentication credentials into the connection request, and opens up a connection to the target service. From there, the Handler simply transparently shuttles data between the client and service.

Select the Handler you are interested in below to learn about its usage and configuration. Are we missing something vital?
Please check our [GitHub issues](https://github.com/cyberark/secretless-broker/issues) to see if the Target Service you are
interested in is on our radar, and request it by opening a GitHub issue if not.

- [AWS](/docs/reference/handlers/http/aws.html)
- [Conjur](/docs/reference/handlers/http/conjur.html)
- [HTTP Basic Authentication](/docs/reference/handlers/http/basic.html)
- [MySQL](/docs/reference/handlers/mysql.html)
- [Postgres](/docs/reference/handlers/postgres.html)
- [SSH Agent](/docs/reference/handlers/ssh_agent.html)
- [SSH](/docs/reference/handlers/ssh.html)
