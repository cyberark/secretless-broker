#!/bin/bash -ex

docker-compose build myapp_tls

docker-compose up --no-deps -d myapp_tls

docker-compose run --rm -v $PWD/src/proxy_tls/proxy_tls.pem:/proxy_tls.pem client curl --cacert /proxy_tls.pem https://myapp_tls
