#!/bin/bash -e

if [[ "$NO_SSL" = "true" ]]
then
  echo "removing SSL support"
  rm /var/lib/mysql/*.pem
  echo "ssl=0" >> /etc/my.cnf
else
  cp /ssl/ca.pem /var/lib/mysql/ca.pem
  cp /ssl/ca-key.pem /var/lib/mysql/ca-key.pem
  cp /ssl/client.pem /var/lib/mysql/client-cert.pem
  cp /ssl/client-key.pem /var/lib/mysql/client-key.pem
  cp /ssl/server.pem /var/lib/mysql/server-cert.pem
  cp /ssl/server-key.pem /var/lib/mysql/server-key.pem
fi
