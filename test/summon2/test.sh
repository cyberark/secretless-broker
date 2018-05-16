#!/bin/bash -e

docker run --rm \
  -v $PWD/../../:/go/src/github.com/conjurinc/secretless \
  -w /go/src/github.com/conjurinc/secretless/test/summon2/ \
  secretless-dev bash -c "
    go test -v .
  "
