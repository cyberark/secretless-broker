#!/bin/bash
#
# Builds secretless sidecar injector mutating webhook service
# usage: ./bin/build
set -ex

docker build -t cyberark/secretless-sidecar-injector:latest .
