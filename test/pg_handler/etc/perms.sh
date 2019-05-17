#!/bin/bash -e

chown postgres:postgres /var/lib/postgresql/server.pem
chown postgres:postgres /var/lib/postgresql/server-key.pem
chown postgres:postgres /var/lib/postgresql/ca.pem
chmod 0600 /var/lib/postgresql/server-key.pem

echo "we did that .____."