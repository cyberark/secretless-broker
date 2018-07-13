# quick-start
An introductory walkthrough of Secretless brokering access to PostgreSQL, SSH 
and HTTP.

## Building and Testing
To build the Docker image:
```
$ ./bin/build
```

To test the Docker image:
```
$ ./bin/test
```

## PostgreSQL Quick-start
1. Download and run the Secretless quick-start as a Docker container:
```
docker container run --rm -p 5454:5454 cyberark/secretless-quickstart
```
2. That's all! Now you can connect to the password-protected PostgreSQL database via the Secretless broker, _without knowing its password_. Give it a try:
```
psql \
  --host localhost \
  --port 5454 \
  --username secretless \
  -d quickstart \
  -c 'select * from counties;'
```

### SSH Quick-start
1. Download and run the Secretless quick-start as a Docker container:
```
docker container run --rm -p 2222:2222 cyberark/secretless-quickstart
```
1. That's all! Now you can establish an SSH connection through the Secretless broker _without knowing any credentials_. Give it a try:
```
ssh -p 2222 user@localhost
```

### HTTP Quick-start
1. Download and run the Secretless quick-start as a Docker container:
```
docker container run --rm -p 80:80 -p 8081:8081 cyberark/secretless-quickstart
```
1. That's all! Now you can make an authenticated HTTP request by using the Secretless broker as a proxy _without knowing any credentials_. Give it a try:
```
http_proxy=localhost:8081 curl -i localhost
```