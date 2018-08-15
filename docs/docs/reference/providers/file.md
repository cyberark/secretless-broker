---
title: Credential Providers
id: file
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/file.html
---

## File
The file provider (`file`) allows you to use a file available to the Secretless Broker
process and/or container as sources of credentials.

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
        provider: file
        id: /run/conjur/conjur-access-token
```
