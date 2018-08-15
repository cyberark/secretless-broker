#!/bin/bash
#
# Builds secretless sidecar injector mutating webhook service
# usage: ./bin/build
set -ex

docker build -f ./sidecar-injector/Dockerfile -t cyberark/secretless-sidecar-injector:latest ./sidecar-injector
