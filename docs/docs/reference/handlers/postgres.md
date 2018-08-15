---
title: Handlers
id: postgres
sub_id: Handlers
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/postgres.html
---

## PostgreSQL
The PostgreSQL handler authenticates incoming connections for a particular
listener.

### Credentials
- `address`  
_Required_  
Connection string of the form `host[:port][/dbname]`  

- `username`  
_Required_  
Username of the PostgreSQL account to connect as  

- `password`  
_Required_  
Password of the PostgreSQL account to connect with  

### Examples
#### Listening on a network address
``` yaml
listeners:
  - name: pg_listener
    protocol: pg
    address: 0.0.0.0:5432

handlers:
  - name: pg_handler
    listener: pg_listener
    credentials:
      - name: address
        provider: literal
        id: postgres.my-service.internal:5432
      - name: username
        provider: literal
        id: my-service
      - name: password
        provider: env
        id: PG_PASSWORD
```
---
#### Listening on a Unix-domain socket
``` yaml
listeners:
  - name: pg_listener
    protocol: pg
    socket: /sock/.s.PGSQL.5432

handlers:
  - name: pg_handler
    listener: pg_listener
    credentials:
      - name: address
        provider: literal
        id: postgres.my-service.internal:5432
      - name: username
        provider: literal
        id: my-service
      - name: password
        provider: env
        id: PG_PASSWORD
```
