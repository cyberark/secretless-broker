FROM google/cloud-sdk:566.0.0-alpine

RUN mkdir -p /src
WORKDIR /src

# Install Docker client
RUN apk add --no-cache ca-certificates curl wget docker-cli

# Install kubectl and GKE auth plugin via gcloud components so the kubectl
# version is guaranteed to be compatible with both gke-gcloud-auth-plugin
# and the GKE cluster managed by this SDK release.
RUN gcloud components install gke-gcloud-auth-plugin kubectl --quiet
