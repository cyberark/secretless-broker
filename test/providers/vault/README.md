# HashiCorp Vault Provider tests

## Local Environment (Laptop)

These instructions show how to develop the HashiCorp Vault (`vault`) Provider on you local machine.

First you'll need a Vault server. It's easy to bring one up using the provided `docker-compose.yml` which will create one in development mode:

```sh-session
$ docker-compose up -d vault
$ # Find the root_token
$ docker-compose logs vault | grep "Root Token:" | head -1 | awk '{print $NF}'
ab5f5b60-5946-e6ff-2d09-c8bf84a08c89
```

Docker has automatically mapped port 8200 of the `vault` container to a local port. Find out this port using `docker-compose port`:

```sh-session
$ docker-compose port vault 8200
0.0.0.0:32812
```

Here's how to add a new secret to the `kv/db/password` variable:

```sh-session
$ # Assumes you have the root token from the vault - you can see it with `docker-compose logs vault`
$ docker-compose run --rm \
    -e VAULT_ADDR=http://vault:8200 \
    -e VAULT_TOKEN="$root_token" \
    --entrypoint vault \
    vault mount kv
Success! Enabled the kv secrets engine at: kv/

$ docker-compose run --rm \
    -e VAULT_ADDR=http://vault:8200 \
    -e VAULT_TOKEN="$root_token" \
    --entrypoint vault \
    vault kv put kv/db/password value=db-secret
Success! Data written to: kv/db/password
```

To read that same value:

```sh-session
$ # Assumes you have the root token from the vault - you can see it with `docker-compose logs vault`
$ docker-compose run --rm \
    -e VAULT_ADDR=http://localhost:32812 \
    -e VAULT_TOKEN="$root_token" \
    --entrypoint vault \
    vault kv get kv/db/password
```

## Testing your local changes

You can test your local changes by re-building the Docker images (running
`./bin/build` in the project root) and then running the test suite as usual:
```
./start
./test
```
or you can run the test suite in "local mode", which will mount your project
directory as a volume in the container, overwriting the version of the project
added to the image in the last build:
```
./start
./test -l
```

## Cleaning up

Invoking the `./stop` script in this directory will clean your workspace up.
