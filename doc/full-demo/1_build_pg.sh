#!/bin/bash -ex

docker-compose build

docker-compose up -d pg ansible_secretless

docker-compose run --rm ansible ansible-playbook ./ansible/pg/postgresql.yml
