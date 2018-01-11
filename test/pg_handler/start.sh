#!/bin/bash -ex

docker-compose build
docker-compose up -d pg secretless
