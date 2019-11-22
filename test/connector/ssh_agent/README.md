# SSH Agent test

This test runs a connection test using the Secretless Broker and a SSH agent socket. Ideally the
test will ensure that connection to a secured ssh-host required no authentication data
from the client connecting over the Secretless Broker.

## Testing

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

## Development

From this directory run `./run_dev` which should get you into a container with Secretless Broker code.

You can start the Secretless Broker with:
```
$ go run cmd/secretless/main.go -f test/ssh_agent_handler/secretless.dev.yml
```

You can test the code by first either:
- running the previous command in the background (appending ` &` to it)
- or connecting with a different terminal to the container with `docker-compose exec dev /bin/bash`

After doing that, you can try opening a separate connection to the Secretless Broker on `/sock/.agent` and ensuring
that no authentication is needed.

_Note: You may need to remove the `/sock/.agent` after you exit the Secretless Broker to be able to start it again_

```
$ export SSH_AUTH_SOCK=/sock/.agent
$ ssh -o StrictHostKeyChecking=no root@ssh_host 'ls -la /'
```

## Cleaning up

From this directory, run `./stop`.
