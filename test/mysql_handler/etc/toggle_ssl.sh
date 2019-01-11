#!/bin/bash -e
# This script is used to build the ROOT/test/mysql_handler mysql container image
# The script expects ROOT/test/util/ssl to contain the pre-generated
# shared SSL fixtures used during testing
#
# This script is housed in /docker-entrypoint-initdb.d/ inside the container image
# The envvar NO_SSL is used to toggle SSL for the mysql container image at startup

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
