FROM google/cloud-sdk:566.0.0-alpine

RUN mkdir -p /src
WORKDIR /src

# Install Docker client
RUN apt-get update -y && \
    apt-get install -y ca-certificates curl wget && \
    install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc && \
    chmod a+r /etc/apt/keyrings/docker.asc && \
    echo "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
      "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update -y && \
    apt-get install -y docker-ce-cli

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
