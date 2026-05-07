FROM google/cloud-sdk:566.0.0-alpine

RUN mkdir -p /src
WORKDIR /src

# Install Docker client
RUN apk add --no-cache ca-certificates curl wget docker-cli

# Required for kubectl auth against GKE clusters
RUN gcloud components install gke-gcloud-auth-plugin --quiet

# Install kubectl CLI
RUN wget -q -O /usr/local/bin/kubectl \
      https://storage.googleapis.com/kubernetes-release/release/v1.11.3/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install kubectx and kubens
RUN wget -q -O /usr/local/bin/kubectx \
      https://raw.githubusercontent.com/ahmetb/kubectx/master/kubectx && \
    chmod +x /usr/local/bin/kubectx
RUN wget -q -O /usr/local/bin/kubens \
      https://raw.githubusercontent.com/ahmetb/kubectx/master/kubens && \
    chmod +x /usr/local/bin/kubens
