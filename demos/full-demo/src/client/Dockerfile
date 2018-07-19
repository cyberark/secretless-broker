FROM ubuntu:18.04

RUN apt-get update && \
    apt-get install -y curl \
                       jq \
                       postgresql-client \
                       vim

COPY . .
