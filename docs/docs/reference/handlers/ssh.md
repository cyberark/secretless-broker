---
title: Secretless Handlers
id: ssh
layout: docs
description: Secretless Documentation
permalink: docs/reference/handlers/ssh
---

# SSH
## Overview
The SSH handler acts as a man-in-the-middle, authenticating inbound SSH 
connections automatically without exposing passwords or keys.

## Credentials
- `address`  
_Required_  
Server address of the form `host[:port]`  

- `privateKey`  
_Required_  
PEM encoded private key  

- `user`  
_Optional_  
User to SSH as (defaults to `root`)  

- `hostKey`  
_Optional_  
accepts any host key if not included  

## Example
``` yaml
listeners:
  - name: ssh_listener
    protocol: ssh
    address: 0.0.0.0:22

handlers:
  - name: ssh_handler
    listener: ssh_listener
    credentials:
      - name: privateKey
        provider: conjur
        id: my-service/ssh-key
      - name: address
        provider: literal
        id: my-service.myorg.com:29341
```