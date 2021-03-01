# certs

This folder contains the cert-key pair for the test server.

These cert-key pairs are self-signed and were generated using an
invocation of openssl similar to this:

```bash
openssl req \
  -x509 \
  -newkey rsa:4096 \
  -keyout server-key.pem \
  -out server-cert.pem \
  -subj '/CN=mismatchedhost' \
  -addext "subjectAltName = DNS:mismatchedhost" \
  -nodes \
  -days 365000
```
