FROM registry2.itci.conjur.net/conjurinc/secretless:latest as secretless
FROM postgres:9.6.9-alpine

MAINTAINER "CyberArk Software, Inc."
LABEL maintainer="CyberArk Software, Inc."

EXPOSE 80 8081 2222 5454
USER root
ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/entrypoint" ]

COPY --from=secretless /usr/local/bin/secretless  /usr/local/bin/
COPY bin/entrypoint                 /
COPY bin/pg-init.sh                 /docker-entrypoint-initdb.d/
COPY etc/nginx.conf                 /etc/nginx/
COPY etc/secretless.yml etc/motd    /etc/

RUN apk add -U openssl openssh nginx apache2-utils tini \
    && mkdir -p /lib64 /etc/nginx /run/nginx /home/user/.ssh/ \
    && ln -fs /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \
    && ssh-keygen -A \
    && adduser -DH secretless \
    && chown secretless /etc/secretless.yml \
    && adduser -s /bin/bash -D user \
    && passwd -u user \
    && sed \
        -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' \
        /etc/ssh/sshd_config
