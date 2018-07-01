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
