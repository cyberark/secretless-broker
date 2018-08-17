---
title: Credential Providers
id: env
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/env.html
---

## Environment
The environment provider (`env`) allows the use of environment variables as a
source of credentials.

### Examples
``` yaml
listeners:
  - name: http_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: conjur_handler
    listener: http_listener
    type: conjur
    match: [ ".*" ]
    credentials:
      - name: accessToken
        provider: env
        id: ACCESS_TOKEN
```
