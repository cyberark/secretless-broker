---
title: Secretless
id: aws
layout: docs
description: Secretless Documentation
permalink: docs/reference/handlers/http/aws
---

# HTTPS AWS Handler
## Overview
The AWS handler exposes an HTTP proxy which will authenticate requests made to
AWS without revealing access keys to the consumer.

## Handler Parameters
- `type`  
_Required_  
This parameter indicates the type of service proxied by the handler. For AWS,
the value of `type` should always be `aws`.  

- `match`  
_Required_  
An array of regex patterns which match a request URI, either partially or fully.
Requests which are matched by a regex in this array will be authenticated by
this handler.  

## Credentials
- `accessKeyID`  
_Required_  
AWS access key ID  

- `secretAccessKey`  
_Required_  
AWS secret access key  

- `accessToken`  
_Required_  
AWS session token  

## Examples
#### Authenticate all requests
``` yaml
listeners:
  - name: http_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: aws_handler
    listener: http_listener
    type: aws
    match:
      - .*
    debug: true
    credentials:
      - name: accessKeyId
        value:
          environment: AWS_ACCESS_KEY_ID
      - name: secretAccessKey
        value:
          environment: AWS_SECRET_ACCESS_KEY
```
---
#### Only authenticate requests to Amazon EC2
``` yaml
listeners:
  - name: http_listener
    protocol: http
    address: 0.0.0.0:8080

handlers:
  - name: aws_handler
    listener: http_listener
    type: aws
    match:
      - ^https\:\/\/ec2\..*\.amazonaws.com
    debug: true
    credentials:
      - name: accessKeyId
        provider: env
        id: AWS_ACCESS_KEY_ID
      - name: secretAccessKey
        provider: env
        id: AWS_SECRET_ACCESS_KEY
```