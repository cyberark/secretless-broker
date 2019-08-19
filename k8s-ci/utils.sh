#!/bin/bash

set -euo pipefail

# Sets additional required environment variables that aren't available in the
# secrets.yml file, and performs other preparatory steps

# Prepare Docker images
function prepareTestEnvironment() {
  # Pipe the Dockerfile into the command to avoid sending the whole
  # context to Docker
  docker build --rm --tag "gke-utils:latest" - < Dockerfile
}

# Delete an image from GCR, unless it is has multiple tags pointing to it
# This means another parallel build is using the image and we should
# just untag it to be deleted by the later job
function deleteRegistryImage() {
  if [ $# -ne 2 ]; then
    echo "ERROR: Usage: deleteRegistryImage <image> <tag>" 1>&2
    return 1
  fi

  local image=$1
  local tag=$2

  runDockerCommand "
    image_digest=\$(gcloud container images list-tags --filter='tags[]=${tag}' --format='get(digest)' '${image}')

    gcloud container images untag -q '${image}:${tag}'

    image_tags=\$(gcloud container images list-tags --filter=digest=\${image_digest} --format='get(tags)' '${image}' | awk -F';' '{print NF}')
    if [ \${image_tags} -eq 0 ]; then
      echo 'No tags left - completely deleting the image...'
      gcloud container images delete -q ${image}@\${image_digest}
    fi
  " > /dev/null
}

function runDockerCommand() {
  docker run --rm \
    -i \
    -e DOCKER_REGISTRY_URL \
    -e DOCKER_REGISTRY_PATH \
    -e GCLOUD_SERVICE_KEY="/tmp${GCLOUD_SERVICE_KEY}" \
    -e GCLOUD_CLUSTER_NAME \
    -e GCLOUD_ZONE \
    -e SECRETLESS_IMAGE \
    -e GCLOUD_PROJECT_NAME \
    -v "${GCLOUD_SERVICE_KEY}:/tmp${GCLOUD_SERVICE_KEY}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$PWD/..":/src \
    -w /src \
    "gke-utils:latest" \
    bash -exc "
      ./k8s-ci/platform_login > /dev/null
      $1
    "
}

function announce() {
  echo "++++++++++++++++++++++++++++++++++++++"
  echo ""
  echo "$@"
  echo ""
  echo "++++++++++++++++++++++++++++++++++++++"
}
