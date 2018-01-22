# Full Demo

## Overview

The purpose of this demo is to show how to use Secretless and Summon2 comprehensively to securely deploy a variety
of applications in a vault-independent way.

## Secretless vs Summon2

Secretless and Summon2 are two different approaches to delivering the secrets needed by an application. 
Secretless is a more secure option, since it hides the secrets from the application code completely. However,
Secretless cannot be used in all cases. For example:

* If an application requires secrets to be present on disk and the application code cannot be modified. Examples: `nginx` with SSL certificates. `postgresql` server with SSL client authentication.
* If the application needs to connect to a backend service that is not yet implemented in Secretless. For example, a client application which connects to a database, and the database protocol is not implemented in Secretless.

## Vault-Independence

Secretless and Summon2 give organizations the flexibility to securely deliver secrets from any vault to applications and infrastructure. In addition, Secretless and Summon2 can utilize multiple vaults simultaneously, and can switch seamlessly between vaults. 

## Demo Components

This demo deploys a custom application behind a reverse proxy which provides SSL termination. The application connects to a Postgresql database.

The application and the reverse proxy run in containers that are linked to each other on the same network. The database runs on a separate machine, and is provisioned by Ansible over SSH. 
