# Conjur Handler and Provider Development

## Local Environment (Laptop)

These instructions show how to develop the Conjur Handler and Provider on you local machine. This way you can use niceties such as IDE features.

First you'll need a Conjur server. It's easy to bring one up using the provided `docker-compose.yml`:

```sh-session
$ docker-compose up -d conjur
```

Next, run `secretless-broker` in a terminal:

```sh-session
$ ./run_dev
2018/01/16 11:06:03 Secretless starting up...
...
2018/01/16 11:06:03 Loaded provider 'conjur'
2018/01/16 11:06:03 http listener 'http_default' listening at: [::]:1080```
```

Docker has automatically mapped port 80 of the `conjur` container to a local port. Find out this port using `docker-compose port`:

```sh-session
$ docker-compose port conjur 80
0.0.0.0:32812
```

Now you can use `secretless-broker` as a forward proxy to Conjur. `secretless-broker` will add the authentication automatically.

Here's how to run `curl`, without any credentials, to list the data in Conjur (replace the port `31812` with the port number you just obtained):

```sh-session
$ http_proxy=http://localhost:1080 curl http://localhost:32812/resources/dev
[{"created_at":"2018-01-16T15:37:37.163+00:00","id":"dev:policy:root","owner":"dev:user:admin","permissions":[],"annotations":[],"policy_versions":[{"version":1,"created_at":"2018-01-16T15:37:37.163+00:00","policy_text":"- !variable db/password\n","policy_sha256":"0528114a50abfa74569eef3d74ac7648fdc440b5b74474593875758b82eb6dd2","id":"dev:policy:root","role":"dev:user:admin"}]},{"created_at":"2018-01-16T15:37:37.163+00:00","id":"dev:variable:db/password","owner":"dev:user:admin","policy":"dev:policy:root","permissions":[],"annotations":[],"secrets":[{"version":1}]}]
```

Here's how to add a new secret to the `db/password` variable:

```sh-session
$ http_proxy=http://localhost:1080 curl -X POST --data secret http://localhost:32812/secrets/dev/variable/db/password
...
2018/01/16 11:16:26 Received response 201 Created
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
