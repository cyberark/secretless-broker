---
title: Handlers
id: mysql
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/handlers/mysql.html
---

## MySQL

The MySQL handler authenticates and brokers connections to a MySQL database.

To secure connections, we support all the MySQL SSL options you're familar
with. See the `sslmode` option below for details.

Note that, unlike most clients, the default `sslmode` for Secretless is
`required`, since nearly all use cases require TLS.  If you do need to turn it
off, however, and know you can do so safely, you can.

### Configuring the Handler

You tell Secretless where to find your database connection details in the yaml
file's `credentials` section.

There you specify where to find your database's address, your username and password,
as well as the `sslmode` details, including the location of any relevant certificates
and revocation lists, if applicable.

The options are as follows:

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

- `sslmode`
  _Optional_

  This option determines if the connection between Secretless and your database
  will be protected by SSL.

  NOTE: As mentioned above, the default is `require` as opposed to `prefer`,
  forcing SSL unless you explicitly turn it off.

  The MySQL documentation website provides detail on the [levels of protection](https://dev.mysql.com/doc/refman/5.7/en/encrypted-connection-options.html)
  provided by different values for the sslmode parameter.

  There are [five modes](https://dev.mysql.com/doc/refman/5.7/en/encrypted-connection-options.html#option_general_ssl-mode):

  + `disable`
  Corresponds to `DISABLED`. Only try a non-SSL connection.

  + `prefer` (not yet supported)
  Corresponds to `PREFERRED`. First try an SSL connection; if that fails, try a non-SSL connection.

  + `require` (default)
  Corresponds to `REQUIRED`. Only try an SSL connection. As is the MySQL standard,
  if a root CA file is present in this mode no verification of the server certificate
  will be done, despite a CA certificate option being specified.

  + `verify-ca`
  Corresponds to `VERIFY_CA`. Only try an SSL connection, and verify that the
  server certificate is issued by a trusted certificate authority (CA).

  + `verify-full` (not yet supported)
  Corresponds to `VERIFY_IDENTITY`. Like `verify-ca`, but additionally perform
  host name identity verification by checking the host name the client uses for
  connecting to the server against the identity in the certificate that the
  server sends to the client.

_NOTE_: If `sslmode` is set to `require`, `verify-ca`, or `verify-full`, it may
be necessary to set some of the values below. The particular values needed
depend on your use case.

+ `sslcert`
  _Optional_

  The content of this parameter specifies the client SSL certificate in PEM format. This
  parameter is ignored if an SSL connection is not made. Corresponds to `ssl-cert`.

+ `sslkey`
  _Optional_

  The content of this parameter specifies the secret key used for the client
  certificate. This parameter is ignored if an SSL connection is not made.
  Corresponds to `ssl-key`.

+ `sslrootcert`
  _Optional_

  The content of this parameter specifies the SSL certificate authority (CA)
  certificate(s) in PEM format. If present, the server's certificate will be
  verified to be signed by one of these authorities. Corresponds to `ssl-ca`.

+ `sslcrl` (not yet supported)
  _Optional_

  This content of this parameter specifies the SSL certificate revocation list
  (CRL) in PEM format. Certificates listed, if present, will be rejected while
  attempting to authenticate the server's certificate. Corresponds to `ssl-crl`.

### Examples

**Listening on a network address with default `sslmode` of `require`**
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
**Listening on a Unix-domain socket with default `sslmode` of `require`**
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

---
**Listening on a network address with verifiable CA, revocation list and private key-pair**
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
      # NOTE: if your CA is stored in the environment
      # or a secret store, rather than a file, you can
      # use the appropriate provider
      - name: sslrootcert
        provider: file
        id: /etc/mysql/root.crt
      - name: sslcert
        provider: file
        id: /etc/mysql/client.crt
      - name: sslkey
        provider: file
        id: /etc/mysql/client.key
      - name: sslcrl
        provider: file
        id: /etc/mysql/root.crl
```
