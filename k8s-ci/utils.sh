#!/bin/bash

set -euo pipefail

# Sets additional required environment variables that aren't available in the
# secrets.yml file, and performs other preparatory steps
function prepareTestEnvironment() {
  # Prepare Docker images
  docker build --rm --tag "gke-utils:latest" - < Dockerfile > /dev/null
}

# Delete an image from GCR, unless it is has multiple tags pointing to it
# This means another parallel build is using the image and we should
# just untag it to be deleted by the later job
function deleteRegistryImage() {
  local image_and_tag=$1

  runDockerCommand "
    gcloud container images delete --force-delete-tags -q ${image_and_tag}
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
    -v ~/.config:/root/.config \
    -v "$PWD/..":/src \
    -w /src \
    "gke-utils:latest" \
    bash -c "
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
