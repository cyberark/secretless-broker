#!/bin/bash -ex

docker-compose build
docker-compose up -d mysql 

./wait-for-mysql

# start secretless once mysql is running
docker-compose up -d secretless

sleep 2
