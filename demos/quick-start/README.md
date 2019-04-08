# quick-start
An introductory walkthrough of the Secretless Broker brokering access to PostgreSQL, SSH
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
1. Download and run the Secretless Broker quick-start as a Docker container:
```
docker container run \
  --rm \
  -p 5432:5432 \
  -p 5454:5454 \
  cyberark/secretless-broker-quickstart
```
2. Direct access to the PostgreSQL database is available over port `5432`. You
can try querying some data, but you don't have the credentials required to
connect:

[//]: # "NOTE: The psql command below uses the universal Keyword/Value Connection Strings, see https://www.postgresql.org/docs/9.2/libpq-connect.html#LIBPQ-CONNSTRING. Do not change to flag-based connection options, they are not universal."
```
psql \
  "host=localhost
  port=5432
  sslmode=disable
  user=secretless
  dbname=quickstart" \
  -c 'select * from counties;'
```
3. The good news is that you don't need any credentials! Instead, you can
connect to the password-protected PostgreSQL database via the Secretless Broker
on port `5454`, _without knowing the password_. Give it a try:

[//]: # "NOTE: The psql command below uses the universal Keyword/Value Connection Strings, see https://www.postgresql.org/docs/9.2/libpq-connect.html#LIBPQ-CONNSTRING. Do not change to flag-based connection options, they are not universal."
```
psql \
  "host=localhost
  port=5454
  sslmode=disable
  user=secretless
  dbname=quickstart" \
  -c 'select * from counties;'
```

### SSH Quick-start
1. Download and run the Secretless Broker quick-start as a Docker container:
```
docker container run \
  --rm \
  -p 2221:22 \
  -p 2222:2222 \
  cyberark/secretless-broker-quickstart
```
2. The default SSH service is exposed over port `2221`. You can try opening an
SSH connection to the server, but you don't have the credentials to log in:
```
ssh -p 2221 user@localhost
```
3. The good news is that you don't need credentials! You can establish an SSH
connection through the Secretless Broker on port `2222` _without any
credentials_. Give it a try:
```
ssh -p 2222 user@localhost
```

### HTTP Quick-start
1. Download and run the Secretless Broker quick-start as a Docker container:
```
docker container run \
  --rm \
  -p 8080:80 \
  -p 8081:8081 \
  cyberark/secretless-broker-quickstart
```
2. The service we're trying to connect to is listening on port `8080`. If you
try to access it, the service will inform you that you're unauthorized:
```
curl -i localhost:8080
```
3. Instead, you can make an authenticated HTTP request by proxying through the
Secretless Broker on port `8081`. The Secretless Broker will inject the proper credentials
into the request _without you needing to know what they are_. Give it a try:
```
http_proxy=localhost:8081 curl -i localhost:8080
```
