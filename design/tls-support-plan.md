# Downstream TLS Support for Database Handlers

To improve the security of the Secretless Broker when opening connections to a database backend, the connection opened between the broker and the target service should happen over SSL.

## General approach
+ Extract sslmode params from credentials into backend connection options
  + sslmode: Provide different levels of protection.
  + sslrootcert: The Certificate Authority (CA) certificate file. This option, if used, must specify the same certificate used by the server
  + sslcert: The client public key certificate file.
  + sslkey: The client private key file.
  + sslcrl: The certificates revoked by certificate authorities.
+ Dial backend to get net.Conn
+ Default to secure connection if TLS supported
+ Ensure client TLS requirements match server requirements and support, FAIL if not.
  + Default to required secured connection, regardless of support
+ Upgrade net.Conn to tls.Conn and use appropriate strategy for sslmode

### Test cases

Given DB container services with TLS support and without:
+ each sslmode works or errors as intended
+ happy path to secretless via unix socket, and tcp

## MySQL
https://dev.mysql.com/doc/refman/8.0/en/using-encrypted-connections.html#using-encrypted-connections-client-side-configuration

### Handshake

https://dev.mysql.com/doc/internals/en/connection-phase.html

Note that throughout the handshake a sequenceID is maintained which is incremented at every step. This means a handshake with TLS-encryption will have a final sequenceID higher than a plain handshake.

1. Upon initial Dial the server responds to client with capabilities
2. Client creates HandshakeResponse with capabilities in common .e.g TLS
3. To initiate an TLS-encrypted connection the client sends an SSLRequest, a truncated version of HandshakeResponse. See https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::SSLRequest
4. Finally, client sends HandshakeResponse which contains authentication information. See https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::HandshakeResponse
5. Deal with Authentication Method Mismatch. See https://dev.mysql.com/doc/internals/en/authentication-method-mismatch.html

### Implementation

NOTE: MySQL can CREATE users [required to use SSL](https://dev.mysql.com/doc/refman/8.0/en/create-user.html#create-user-tls)

support for "sslmode" = (
  disable
  allow
  prefer
  require
  verify-ca
  verify-full
)

** NOTE - as a mysql user you might be used to the following notation but we only support the notation above **

support for "ssl-mode" = (
  DISABLED
  PREFFERED
  REQUIRED
  VERIFY_CA
  VERIFY_IDENTITY
)

### OSS libraries

+ https://github.com/go-sql-driver/mysql/blob/master/packets.go#L322-L335

## Postgres:
https://www.postgresql.org/docs/9.1/libpq-ssl.html#LIBPQ-SSL-PROTECTION

### Handshake
https://www.postgresql.org/docs/9.3/protocol-flow.html

1. Upon initial Dial the server waits for the client to issue a StartupMessage; This message includes the names of the user and of the database the user wants to connect to; it also identifies the particular protocol version to be used.
2. To initiate an TLS-encrypted connection, the frontend initially sends an SSLRequest message rather than a StartupMessage
3. The client sends the StartupMessage
4. The server then sends an appropriate authentication request message, to which the client must reply with an appropriate authentication response message.

### Implementation

support for "sslmode" = (
  disable
  allow
  prefer
  require
  verify-ca
  verify-full
)

### OSS libraries

+ https://github.com/lib/pq/blob/master/ssl.go
+ https://github.com/CrunchyData/crunchy-proxy/tree/master/connect


##  Stories

Each handler should have a corresponding story

+ [ ] handler documentation provides example and descriptions of sslmode params
      in credentials (see [example](tls-support-plan.md#example-updated-postgresql-documentation))

+ [ ] handler supports TLS defaulting to sslmode=require

  A.C

  + When a connection is made to a server:
    + FAIL, if the server does not support TLS
    + DO NOT VERIFY server certificate, otherwise
  + test cases exist for each of the scenarios above

+ [ ] handler supports sslrootcert and sslmode up to verify-ca

  A.C

  + handler credentials accepts sslmode and sslrootcert
  + [default] When a connection is made to a server:
    + FAIL, if the server does not support TLS
    + DO NOT VERIFY, if no root CA is present
    + VERIFY the server certificate (same as verify-ca), if a root CA file is present
  + When a connection is made to a server:
    + FAIL, if client requires TLS and server does not support it
    + DO NOT USE TLS, if sslmode=disable
    + ONLY USE TLS, if sslmode=prefer and server supports TLS
  + test cases exist for each of the scenarios above

+ [ ] handler supports sslmode=verify-full

  A.C

  + When a connection is made to a server:
    + FAIL, if the server cannot verify cert and hostnea
    + SUCCESS, otherwise
  + test cases exist for each of the scenarios above

+ [ ] handler supports private-key pair as sslkey and sslcert

  A.C

  + handler credentials accepts `sslkey` and `sslcert`
  + When a connection is made to a server:
    + FAIL, if the server cannot verify private-key pair
    + SUCCESS, otherwise
  + test cases exist for each of the scenarios above

## Example updated PostgreSQL Documentation

## PostgreSQL

The PostgreSQL handler authenticates and brokers connections to a postgres
database.

To secure connections, we support all the postgres SSL options you're familar
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

  The Postgres documentation website provides detail on the [levels of protection](https://www.postgresql.org/docs/9.1/libpq-ssl.html#LIBPQ-SSL-PROTECTION)
  provided by different values for the sslmode parameter.

  There are [six modes](https://www.postgresql.org/docs/9.1/libpq-connect.html#LIBPQ-CONNECT-SSLMODE):

  + `disable`
  only try a non-SSL connection

  + `allow`
  first try a non-SSL connection; if that fails, try an SSL connection

  + `prefer`
  first try an SSL connection; if that fails, try a non-SSL connection

  + `require` (default)
  only try an SSL connection. If a root CA file is present, verify the certificate in the same way as if verify-ca was specified

  + `verify-ca`
  only try an SSL connection, and verify that the server certificate is issued by a trusted certificate authority (CA).

  + `verify-full`
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

+ `sslcrl`
  _Optional_

  This content of this parameter specifies the SSL certificate revocation list
  (CRL), replacing the default ~/.postgresql/root.crl. Certificates listed, if
  present, will be rejected while attempting to authenticate the server's
  certificate.

### Examples

#### Listening on a network address with default `sslmode` of `require`
``` yaml
version: "2"
services:
  pg_connector:
    connector: pg
    listenOn: tcp://0.0.0.0:5432
    credentials:
      host: postgres.my-service.internal
      username: myservice
      password:
        from: env
        get: PG_PASSWORD
```
---
#### Listening on a network address with verifiable CA, revocation list and private key-pair
``` yaml
version: "2"
services:
  pg_connector:
    connector: pg
    listenOn: tcp://0.0.0.0:5432
    credentials:
      host: postgres.my-service.internal
      username: myservice
      password:
        from: env
        get: PG_PASSWORD
      sslmode: verify-full
      # NOTE: if your CA is stored in the environment
      # or a secret store, rather than a file, you can
      # use the appropriate provider
      sslrootcert:
        from: file
        get: /etc/pg/root.crt
      sslcert:
        from: file
        get: /etc/pg/client.crt
      sslkey:
        from: file
        get: /etc/pg/client.key
      sslcrl:
        from: file
        get: /etc/pg/root.crl
```
---
#### Listening on a Unix-domain socket
``` yaml
version: "2"
services:
  pg_connector:
    connector: pg
    listenOn: unix:///sock/.s.PGSQL.5432
    credentials:
      host: postgres.my-service.internal
      username: myservice
      password:
        from: env
        get: PG_PASSWORD
```
