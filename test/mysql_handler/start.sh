#!/bin/bash -ex

mysql_host_port=$(docker-compose port mysql 3306)
mysql_port=$(echo "$mysql_host_port" | go run ../util/parse_port.go)

export MYSQL_PORT=$mysql_port

docker-compose build
docker-compose up -d mysql secretless

sleep 2
