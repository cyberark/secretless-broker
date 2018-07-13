FROM secretless:latest as secretless

FROM alpine:3.8

RUN apk add --update ansible \
                     libc6-compat \
                     python

COPY --from=secretless /usr/local/bin/summon2 /usr/local/bin/summon2

COPY group_vars ./group_vars
COPY postgresql.yml ./
COPY hosts /etc/ansible/
