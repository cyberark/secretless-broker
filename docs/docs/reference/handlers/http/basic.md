---
title: Secretless Handlers
id: basic
layout: docs
description: Secretless Documentation
permalink: docs/reference/handlers/http/basic
---

## Basic Authentication - HTTP(S)
### Overview
The basic authentication handler exposes an HTTP proxy which will authenticate
requests made to an arbitrary service requiring basic authentication.

### Handler Parameters
- `type`  
_Required_  
This parameter indicates the type of service proxied by the handler. For AWS,
the value of `type` should always be `basic_auth`.  

- `match`  
_Required_  
An array of regex patterns which match a request URI, either partially or fully.
Requests which are matched by a regex in this array will be authenticated by
this handler.  

### Credentials
- `username`  
_Required_  
Username to authenticate with  

- `password`  
_Required_  
Password to authenticate with  

- `forceSSL`  
_Optional_  
Boolean; Forces connection over https if true  

### Examples
#### Authenticates all requests proxied through this handler
``` yaml
listeners:
  - name: basic_auth_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: basic_auth_handler
    listener: basic_auth_listener
    type: basic_auth
    match:
      - .*
    credentials:
      - name: username
        provider: literal
        id: automation
      - name: password
        provider: env
        id: BASIC_AUTH_PASSWORD
```
---
#### Authenticate requests to a particular hostname
``` yaml
listeners:
  - name: basic_auth_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: basic_auth_handler
    listener: basic_auth_listener
    type: basic_auth
    match:
      - ^https\:\/\/password-protected.myorg.com\/.*
    credentials:
      - name: username
        provider: literal
        id: automation
      - name: password
        provider: env
        id: BASIC_AUTH_PASSWORD
```