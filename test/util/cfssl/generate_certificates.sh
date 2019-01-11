#!/usr/bin/env bash
#
# This script was used to generate the shared ssl fixtures in
# ROOT/test/util/ssl
#
# cfssl - Cloudflare's PKI and TLS toolkit - is the utility used to
# generate the ssl fixtures.
# cfssl is available at https://github.com/cloudflare/cfssl
#
# Below we generate the private keys and certificates for
# root ca, server and client
#
# ca-config.json and ca-csr.json, in the same directory as this script, are used
# to configure the generation of the CA certificate.
#

# generate root ca private key and certificate
cfssl gencert -initca ca-csr.json | cfssljson -bare ca -

# generate server private key and root ca signed certificate
echo '
{
  "CN": "server",
  "hosts": [
    ""
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  }
}
' | cfssl gencert \
 -ca=ca.pem \
 -ca-key=ca-key.pem \
 -config=ca-config.json \
 -profile=server \
 -hostname="" \
 - | cfssljson -bare server

# generate client private key and root ca signed certificate
echo '
{
  "CN": "client",
  "hosts": [
    ""
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  }
}
' | cfssl gencert \
 -ca=ca.pem \
 -ca-key=ca-key.pem \
 -config=ca-config.json \
 -profile=client \
 -hostname="" \
 - | cfssljson -bare client
