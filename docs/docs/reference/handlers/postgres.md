---
title: Handlers
id: postgres
description: Secretless Broker Documentation
permalink: docs/reference/handlers/postgres.html
---

## PostgreSQL

The PostgreSQL handler authenticates and brokers connections to a PostgreSQL
database.

To secure connections, we support all the PostgreSQL SSL options you're familar
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

- `address`
  _Required_

  Connection string of the form `host[:port][/dbname]`  

- `username`
  _Required_

  Username of the PostgreSQL account you are connecting as.

- `password`
  _Required_

  Password of the PostgreSQL account you are connecting as.

- `sslmode`
  _Optional_

  This option determines if the connection between Secretless and your database
  will be protected by SSL.

  NOTE: As mentioned above, the default is `require` as opposed to `prefer`,
  forcing SSL unless you explicitly turn it off.

  The PostgreSQL documentation website provides detail on the [levels of protection](https://www.postgresql.org/docs/9.1/libpq-ssl.html#LIBPQ-SSL-PROTECTION)
  provided by different values for the sslmode parameter.

  There are [six modes](https://www.postgresql.org/docs/9.1/libpq-connect.html#LIBPQ-CONNECT-SSLMODE):

  + `disable`
  only try a non-SSL connection

  + `allow` (not yet supported)
  first try a non-SSL connection; if that fails, try an SSL connection

  + `prefer` (not yet supported)
  first try an SSL connection; if that fails, try a non-SSL connection

  + `require` (default)
  only try an SSL connection. If a root CA file is present, verify the certificate in the same way as if verify-ca was specified

  + `verify-ca`
  only try an SSL connection, and verify that the server certificate is issued by a trusted certificate authority (CA).

  + `verify-full` (not yet supported)
  only try an SSL connection, verify that the server certificate is issued by a trusted CA and that the server host name matches that in the certificate

_NOTE_: If `sslmode` is set to `require`, `verify-ca`, or `verify-full`, it may
be necessary to set some of the values below. The particular values needed
depend on your use case.

+ `sslcert`
  _Optional_

  The content of this parameter specifies the client SSL certificate, replacing
  the default ~/.postgresql/postgresql.crt. This parameter is ignored if an SSL
  connection is not made.

+ `sslkey`
  _Optional_

  The content of this parameter specifies the secret key used for the client
  certificate, replacing the default ~/.postgresql/postgresql.key. This
  parameter is ignored if an SSL connection is not made.

+ `sslrootcert`
  _Optional_

  The content of this parameter specifies the SSL certificate authority (CA)
  certificate(s), replacing the default ~/.postgresql/root.crt. If present, the
  server's certificate will be verified to be signed by one of these
  authorities.

+ `sslcrl` (not yet supported)
  _Optional_

  This content of this parameter specifies the SSL certificate revocation list
  (CRL), replacing the default ~/.postgresql/root.crl. Certificates listed, if
  present, will be rejected while attempting to authenticate the server's
  certificate.

### Examples

**Listening on a network address with default `sslmode` of `require`**
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
**Listening on a Unix-domain socket with default `sslmode` of `require`**
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
---
**Listening on a network address with verifiable CA, revocation list and private key-pair**
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
      - name: sslmode
        provider: literal
        id: verify-full
      # NOTE: if your CA is stored in the environment
      # or a secret store, rather than a file, you can
      # use the appropriate provider
      - name: sslrootcert
        provider: file
        id: /etc/pg/root.crt
      - name: sslcert
        provider: file
        id: /etc/pg/client.crt
      - name: sslkey
        provider: file
        id: /etc/pg/client.key
      - name: sslcrl
        provider: file
        id: /etc/pg/root.crl
```
