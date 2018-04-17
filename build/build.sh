#!/bin/bash
#
# Builds secretless binaries
# usage: ./build/build.sh [OPTIONAL LIST OS] [OPTIONAL LIST ARCH]
# If OS/arch arguments are unspecified, builds binaries for all supported 
# operating systems and architectures. Set KEEP_ALIVE=1 to keep the build
# container running after this script exits.
set -ex

on_error () {
  docker rm -f "${CONTAINER_ID}" >/dev/null
  trap - EXIT
}
trap on_error ERR

on_exit () {
  [ -n "${KEEP_ALIVE}" ] || docker rm -f "${CONTAINER_ID}" >/dev/null
}
trap on_exit EXIT

readonly GOLANG_VERSION="1.9"
readonly SUPPORTED_BUILD_OS="windows darwin linux"
readonly SUPPORTED_BUILD_ARCH="386 amd64"
readonly REPOSITORY="github.com/conjurinc/secretless"
readonly CONTAINER_ID="$(docker run \
  -v "$(pwd):/go/src/${REPOSITORY}" \
  -itd \
  golang:${GOLANG_VERSION} \
  bash
)"

BUILD_OS="${1:-${SUPPORTED_BUILD_OS}}"
BUILD_ARCH="${2:-${SUPPORTED_BUILD_ARCH}}"

docker exec "${CONTAINER_ID}" bash -c "
  cd /go/src/${REPOSITORY}
  curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
  dep ensure

  for GOOS in ${BUILD_OS}; do
    for GOARCH in ${BUILD_ARCH}; do
      export GOOS
      export GOARCH
      echo \"Building binaries for \${GOOS} \${GOARCH}\"
      mkdir -p bin/\${GOOS}/\${GOARCH}
      go build -o bin/\${GOOS}/\${GOARCH}/secretless ./cmd/secretless
      go build -o bin/\${GOOS}/\${GOARCH}/summon2 ./cmd/summon2
    done
  done
"

echo "Building docker-compose services"
docker-compose -f build/docker-compose.yml build