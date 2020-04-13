FROM secretless-broker:latest as secretless-broker

FROM postgres:9.6.9-alpine

MAINTAINER "CyberArk Software, Inc."
LABEL maintainer="CyberArk Software, Inc."

EXPOSE 80 8081 2222 5454
USER root
ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/entrypoint" ]

COPY bin/entrypoint                 /

COPY bin/pg-init.sh                 /docker-entrypoint-initdb.d/
COPY etc/pg_server.key              /var/lib/postgresql/server.key
COPY etc/pg_server.crt              /var/lib/postgresql/server.crt

COPY etc/nginx.conf                 /etc/nginx/

COPY etc/secretless.yml etc/motd    /etc/

RUN apk update && apk upgrade

RUN apk add --no-cache apache2-utils \
                       libc6-compat \
                       nginx \
                       tini \
                       openssh \
                       openssl && \
     mkdir -p /lib64 /etc/nginx /run/nginx /home/user/.ssh/ && \
     ssh-keygen -A && \
     adduser -DH secretless && \
     chown secretless /etc/secretless.yml && \
     adduser -s /bin/bash -D user && \
     passwd -u user && \
     sed \
        -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' \
        /etc/ssh/sshd_config && \
    # Go DNS resolution doesn't read /etc/hosts by default. See
    # for more info: https://github.com/golang/go/issues/22846
     echo "hosts: files dns" > /etc/nsswitch.conf && \
     chown postgres:postgres /var/lib/postgresql/server.key && \
     chown postgres:postgres /var/lib/postgresql/server.crt && \
     chmod 0600 /var/lib/postgresql/server.key

COPY --from=secretless-broker /usr/local/bin/secretless-broker /usr/local/bin/
