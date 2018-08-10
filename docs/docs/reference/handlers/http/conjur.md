---
title: Handlers
id: conjur
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/http/conjur.html
---

## Conjur - HTTP(S)
### Overview
The Conjur handler exposes an HTTP proxy which will authenticate requests made
to Conjur without revealing credentials to the consumer.

### Handler Parameters
- `type`  
_Required_  
This parameter indicates the type of service proxied by the handler. For AWS,
the value of `type` should always be `conjur`.  

- `match`  
_Required_  
An array of regex patterns which match a request URI, either partially or fully.
Requests which are matched by a regex in this array will be authenticated by
this handler.  

### Credentials
- `accessToken`  
_Required_  
Conjur access token  

- `forceSSL`  
_Optional_  
Boolean; forces connection over HTTPS if true  

### Examples
#### Authenticates all requests proxied through this handler
``` yaml
listeners:
  - name: http_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: conjur_handler
    listener: http_listener
    type: conjur
    match:
      - .*
    credentials:
      - name: accessToken
        provider: file
        id: /run/conjur/conjur-access-token
```
---
#### Authenticate requests to a particular hostname
``` yaml
listeners:
  - name: http_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: conjur_handler
    listener: http_listener
    type: conjur
    match:
      - ^https\:\/\/conjur.myorg.com\/.*
    credentials:
      - name: accessToken
        provider: file
        id: /run/conjur/conjur-access-token
```
