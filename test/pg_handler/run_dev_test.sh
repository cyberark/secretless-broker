#!/bin/bash -e

./run_dev.sh &

sleep 2

go test .

kill "$!"
