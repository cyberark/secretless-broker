# certs

This folder contains the cert-key pair for the test server.

These cert-key pairs are self-signed and were generated using an
invocation of openssl similar to this:

```bash
function gen_cert() {
  local suffix=$1

  openssl req \
  -x509 \
  -newkey rsa:4096 \
  -keyout server-key-${suffix}.pem \
  -out server-cert-${suffix}.pem \
  -subj '/CN=test' \
  -addext "subjectAltName = DNS:test" \
  -nodes \
  -days 365000
}

gen_cert excluded
gen_cert included
```
