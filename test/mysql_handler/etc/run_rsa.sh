#!/bin/bash -e

if [[ "$NO_SSL" = "true" ]]
then
  echo "removing SSL support"
  rm /var/lib/mysql/*.pem
  echo "ssl=0" >> /etc/my.cnf
fi
