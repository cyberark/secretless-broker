#!/bin/bash -ex

for dir in plaintext conjur; do
  (
    cd $dir
    docker-compose build
  )
done
