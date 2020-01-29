# MSSQL Server Connector

**NOTE: This connector is in beta.**

The MSSQL Server Connector enables:

- Kubernetes or OpenShift applications to connect to a MSSQL Server 2017 v14.x
  database using SQL Server authentication
- Connecting applications to have no knowledge of the required database
  credentials
- New connections to MSSQL to always use fresh credentials from the configured
  credential provider

## Supported configuration (with samples)

Here's a sample `secretless.yml` configuration file that enables Secretless to
connect to an MSSQL Server:

```yaml
version: 2

services:
  mssql:
    connector: mssql
    listenOn: tcp://0.0.0.0:2223
    credentials:
      username: sa
      password:
        from: conjur
        get: my-sql-server-password
      host: mssql
      port: 1433
```

### Required credentials

All credentials are required unless they're explicitly marked as "optional".

- `username` - Database username under SQL Server authentication mode
- `password` - Database password under SQL Server authentication mode
- `host` - The network address of the target MSSQL Server.  In the example
  above, sine we're using Docker networking, it happens to be just the bare
  name `mssql`.
- (optional) `port` - The port the target MSSQL Server is listening on.
  Defaults to the standard 1433.

## Target Service SSL Support

SSL is currently not supported in the beta version, but will be coming soon.

## Supported versions

The connector supports MSSQL Server 2017.

In particular, it is tested against the
`mcr.microsoft.com/mssql/server:2017-latest` [Linux docker
image](https://hub.docker.com/_/microsoft-mssql-server), whose version at the
time of this writing is:  

```
Microsoft SQL Server 2017 (RTM-CU17) (KB4515579) - 14.0.3238.1 (X64)  
Sep 13 2019 15:49:57  
Copyright (C) 2017 Microsoft Corporation 
Developer Edition (64-bit) on Linux (Ubuntu 16.04.6 LTS)
```

## Known limitations

- Does not currently support SSL in the connection between Secretless and MSSQL
  Server.
- Only supports SQL Server authentication mode
- Only limited tests have been performed.  Specifically, the ability to connect
  using Secretless has been tested using two clients:
    - The `sqlcmd` tool that ships with the above version of MSSQL server
    - The Go MSSQL driver provided by the package
      `github.com/denisenkom/go-mssqldb`
- Since we use the `go-mssqldb` package, Secretless is also affected by the
  [known issues of that
  package](https://github.com/denisenkom/go-mssqldb#known-issues).



