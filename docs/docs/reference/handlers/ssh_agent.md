---
title: Handlers
id: ssh_agent
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/ssh_agent.html
---

## SSH Agent
### Overview
The SSH Agent handler enables the Secretless Broker to replace `ssh-agent` by providing
similar functionality over a socket without exposing keys. Once running, export
`SSH_AUTH_SOCK` to equal the path of your listener socket targeted by this
handler.

### Credentials
- `rsa` or `ecdsa`  
_Required_  
RSA or ECDSA private key

- `comment`  
_Optional_  
free-form string  

- `lifetime`  
_Optional_  
if not 0, the number of seconds the agent will store the key for  

- `confirm`  
_Optional_  
confirms with user before using if true  

### Example
``` yaml
listeners:
  - name: ssh_agent_listener
    protocol: ssh-agent
    socket: /sock/.agent

handlers:
  - name: ssh_agent_handler
    listener: ssh_agent_listener
    credentials:
      - name: rsa
        provider: file
        id: /id_rsa
```
  
With the Secretless Broker running this configuration, use it in replacement of
`ssh-agent` by exporting `SSH_AUTH_SOCK`:  
``` bash
$ export SSH_AUTH_SOCK=/sock/.agent
```
