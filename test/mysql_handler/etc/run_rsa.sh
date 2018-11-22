#!/bin/bash -e

mysql_ssl_rsa_setup

cat << EOF > /etc/mysql/mysql.conf.d/ssl.cnf
[mysqld]
ssl-ca=/var/lib/mysql/ca.pem
ssl-cert=/var/lib/mysql/server-cert.pem
ssl-key=/var/lib/mysql/server-key.pem
EOF
