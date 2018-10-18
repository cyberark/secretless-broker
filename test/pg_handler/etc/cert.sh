#!/bin/bash

BASEDIR=$(dirname "$0")
LOGGING_PREFIX="gen_cert.sh >> "

PASSKEY=somekey

rm -f ${BASEDIR}/server.crt
rm -f ${BASEDIR}/server.csr
rm -f ${BASEDIR}/server.key
rm -f ${BASEDIR}/rootCA.crt
rm -f ${BASEDIR}/rootCA.csr
rm -f ${BASEDIR}/rootCA.key
rm -f ${BASEDIR}/rootCA.srl

# generate a key for our root CA certificate
echo "${LOGGING_PREFIX} Generating key for root CA certificate"
openssl genrsa -des3 -passout pass:${PASSKEY} -out ${BASEDIR}/rootCA.pass.key 2048
openssl rsa -passin pass:${PASSKEY} -in ${BASEDIR}/rootCA.pass.key -out ${BASEDIR}/rootCA.key
rm ${BASEDIR}/rootCA.pass.key
echo

# create and self sign the root CA certificate
echo
echo "${LOGGING_PREFIX} Creating self-signed root CA certificate"
openssl req -x509 -new -nodes -key ${BASEDIR}/rootCA.key -sha256 -days 1024 -out ${BASEDIR}/rootCA.crt -subj "/C=UK/ST=/L=/O=IBM/OU=AIOS/CN=aios-orch-dev-env-CA"
echo "${LOGGING_PREFIX} Self-signed root CA certificate (${BASEDIR}/rootCA.crt) is:"
openssl x509 -in ${BASEDIR}/rootCA.crt -text -noout
echo

# generate a key for our server certificate
echo
echo "${LOGGING_PREFIX} Generating key for server certificate"
openssl genrsa -des3 -passout pass:${PASSKEY} -out ${BASEDIR}/server.pass.key 2048
openssl rsa -passin pass:${PASSKEY} -in ${BASEDIR}/server.pass.key -out ${BASEDIR}/server.key
rm ${BASEDIR}/server.pass.key
echo

# create a certificate request for our server. This includes a subject alternative name so either localhost or pg can be used to address it
echo
echo "${LOGGING_PREFIX} Creating server certificate"
openssl req -new -key ${BASEDIR}/server.key -out ${BASEDIR}/server.csr -subj "/C=UK/ST=/L=/O=IBM/OU=AIOS/CN=pg" -reqexts SAN -config <(cat /etc/ssl/openssl.cnf <(printf "[SAN]\nsubjectAltName=DNS:pg,DNS:localhost"))
echo "${LOGGING_PREFIX} Server certificate signing request (${BASEDIR}/server.csr) is:"
openssl req -verify -in ${BASEDIR}/server.csr -text -noout
echo

# use our CA certificate and key to create a signed version of the server certificate
echo
echo "${LOGGING_PREFIX} Signing server certificate using our root CA certificate and key"
openssl x509 -req -sha256 -days 365 -in ${BASEDIR}/server.csr -CA ${BASEDIR}/rootCA.crt -CAkey ${BASEDIR}/rootCA.key -CAcreateserial -out ${BASEDIR}/server.crt -extensions SAN -extfile <(cat /etc/ssl/openssl.cnf <(printf "[SAN]\nsubjectAltName=DNS:pg,DNS:localhost"))
chmod og-rwx ${BASEDIR}/server.key
echo "${LOGGING_PREFIX} Server certificate signed with our root CA certificate (${BASEDIR}/server.crt) is:"
openssl x509 -in ${BASEDIR}/server.crt -text -noout
echo

# done output the base64 encoded version of the root CA certificate which should be added to trust stores
echo
echo "${LOGGING_PREFIX} Done. Next time the pg docker image is rebuilt the new server certificate (${BASEDIR}/server.crt) will be used."
echo
echo "${LOGGING_PREFIX} Use the following CA certificate variables:"
B64_CA_CERT=`cat ${BASEDIR}/rootCA.crt | base64`
echo "pg_CA_CERT=${B64_CA_CERT}"
