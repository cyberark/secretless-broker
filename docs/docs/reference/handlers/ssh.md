---
title: Secretless
id: ssh
layout: docs
description: Secretless Documentation
permalink: docs/reference/handlers/ssh
---

# SSH Handler
## Overview
The SSH handler authenticates incoming SSH connections for a particular
listener without exposing passwords or keys to the consumer.

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