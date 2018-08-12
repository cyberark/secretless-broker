---
title: Handlers
id: mysql
sub_id: Handlers
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/mysql.html
---

## MySQL
### Overview
The MySQL handler authenticates incoming connections for a particular
listener.

### Credentials
- `host`  
_Required_  
Host name of the MySQL server  

- `port`  
_Required_  
Port of the MySQL server  

- `username`  
_Required_  
Username of the MySQL account to connect as  

- `password`  
_Required_  
Password of the MySQL account to connect with  

### Examples
#### Listening on a network address
``` yaml
listeners:
  - name: mysql_listener
    protocol: mysql
    address: 0.0.0.0:3306

handlers:
  - name: mysql_handler
    listener: mysql_listener
    credentials:
      - name: host
        provider: literal
        id: mysql.my-service.internal
      - name: port
        provider: literal
        id: 3306
      - name: username
        provider: literal
        id: my-service
      - name: password
        provider: env
        id: MYSQL_PASSWORD
```
---
#### Listening on a Unix-domain socket
``` yaml
listeners:
  - name: mysql_listener
    protocol: mysql
    socket: /sock/mysql.sock

handlers:
  - name: mysql_handler
    listener: mysql_listener
    credentials:
      - name: host
        provider: literal
        id: mysql.my-service.internal
      - name: port
        provider: literal
        id: 3306
      - name: username
        provider: literal
        id: my-service
      - name: password
        provider: env
        id: MYSQL_PASSWORD
```
