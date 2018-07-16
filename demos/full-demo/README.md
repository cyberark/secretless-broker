# Full Demo

## Overview

The purpose of this demo is to show how to use Secretless and Summon2 comprehensively to securely deploy a variety of applications in a vault-independent way.

## Secretless vs Summon2

Secretless and Summon2 are two different approaches to delivering the secrets needed by an application. Secretless is a more secure option, since it hides the secrets from the application code completely. However, Secretless cannot be used in all cases. For example:

* If an application requires secrets to be present on disk and the application code cannot be modified. Examples: `nginx` with SSL certificates. `postgresql` server with SSL client authentication.
* If the application needs to connect to a backend service that is not yet implemented in Secretless. For example, a client application which connects to a database, and the database protocol is not implemented in Secretless.

## Vault-Independence

Secretless and Summon2 give organizations the flexibility to securely deliver secrets from any vault to applications and infrastructure. In addition, Secretless and Summon2 can utilize multiple vaults simultaneously, and can switch seamlessly between vaults. 

## Building and Running

First you need to build the project, from the top-level project directory:

```sh-session
secretless $ ./bin/build
```

Then from this directory, build each scenario:

```sh-session
secretless/demos/full-demo $ ./build
```

Next, enter the scenario directory. For example, "plaintext":

```sh-session
secretless/demos/full-demo $ cd plaintext
secretless/demos/full-demo/plaintext $
```

In this directory you'll find a sequence of shell scripts. Run each script in sequence.

```sh-session
plaintext $ ./1_build_pg
...
```

```sh-session
plaintext $ ./2_build_myapp
...
```

## Demo Components

In the first flow, Ansible is used to configure a Postgresql database:

`ansible -> pg`

In the second flow, an application "myapp" can create and retrieve user info from the database:

`myapp -> pg`

In the third flow, a proxy providing SSL termination runs in front of "myapp":

`myapp_tls -> myapp -> pg`

### Database `pg`

The database runs in a container which is set up like a VM, with an init process running an SSH daemon as well as postgresql.

A public key is installed to `/root/.ssh/authorized_keys`.

### Ansible

An ansible playbook "postgresql.yml" is used to create a database "myapp", and create a user "myapp" with access to the database. The password for "myapp" is provided by an ansible variable "dbpassword". 

The variable "dbpassword" obtains its value from the environment variable "DB_PASSWORD"; this is setup using an Ansible `group_vars` file.

### Myapp

The app interprets two environment variables: `DB_HOST` and `DB_PASSWORD`. `DB_HOST` can either be a hostname or a path to a Unix domain socket. These variables are used to connect to the database.

### Proxy_tls

This app expects two environment variables: `SSL_CERT_FILE` and `SSL_KEY_FILE`. The app also expects a single command-line argument, which is a "host:port". 

`proxy_tls` runs a TLS listener and provides reverse proxying and SSL termination to the specified "host:port".

## Secrets in Plaintext

### Ansible

* **SSH key file** Provided as a volume-mounted file `/root/id_insecure`.
* **DB_PASSWORD** Provided as a container environment variable.

### Myapp

* **DB_HOST** `pg:5432`
* **DB_PASSWORD** Provided as a container environment variable.

### Proxy_tls

* **SSL_CERT_FILE** The certificate is built into the container.
* **SSL_KEY_FILE** Provided as a volume-mounted file `/proxy_tls.key`.

## Secrets in Conjur

### Ansible

* **SSH key file** Secretless stores the SSH key in its ssh-agent implementation, and the ssh-agent Unix domain socket is shared with Ansible and exposed as `SSH_AUTH_SOCK`.
* **DB_PASSWORD** `summon2` is used as the entrypoint to the container.  `CONJUR_AUTHN_API_KEY` is provided as a container environment variable.

### Myapp

* **DB_HOST** Secretless runs a `pg` listener on Unix socket `/sock/s.PGSQL.5432`. This socket file is shared with `myapp`. `myapp` connects to `pg` through Secretless using this socket.
* **DB_PASSWORD** Not provided.

### Proxy_tls

* **SSL_CERT_FILE** The certificate is built into the container.
* **SSL_KEY_FILE**  `summon2` is used as the entrypoint to the container.  `CONJUR_AUTHN_API_KEY` is provided as a container environment variable. The SSL key is stored by `summon2` as a temp file in the container, which is deleted when the `summon2` exits.

