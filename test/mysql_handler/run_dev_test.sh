#!/bin/bash -ex

function finish {
  rm -rf run/mysql/*
  kill "$!"
}
trap finish EXIT

./run_dev.sh &

# wait for secretless / handler to be ready
while [ ! -S $PWD/run/mysql/mysql.sock ]; do sleep 1; done 2> /dev/null

go test -v .
