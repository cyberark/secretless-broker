FROM alpine:3.7
MAINTAINER CyberArk Software, Inc.

RUN mkdir -p /lib64 \
    && ln -fs /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY ./bin/linux/amd64/secretless /
COPY ./bin/linux/amd64/summon2 /

WORKDIR /

ENTRYPOINT [ "./secretless" ]
