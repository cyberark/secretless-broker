#!/bin/bash -ex

docker-compose run --rm test \
  env SSH_AUTH_SOCK=/run/ssh-agent/.agent ssh -o StrictHostKeyChecking=no ssh_host \
  cat /root/.ssh/authorized_keys
