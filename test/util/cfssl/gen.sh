#!/usr/bin/env bash
#
# This file was used to generate files in
# test/mysql_handler/ssl
# test/pg_handler/ssl

# generate ca
cfssl gencert -initca ca-csr.json | cfssljson -bare ca -

# generate server set
echo '{"CN":"server","hosts":[""],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server -hostname="" - | cfssljson -bare server

# generate client set
echo '{"CN":"client","hosts":[""],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client -hostname="" - | cfssljson -bare client2
