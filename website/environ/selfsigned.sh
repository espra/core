#! /usr/bin/env bash

# Public Domain (-) 2017-present, The Core Authors.
# See the Core UNLICENSE file for details.

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

trap log_error ERR

function log_error() {
    printf "\n\033[31;1m!! ERROR:\n!! ERROR: failed to run ./selfsigned.sh\n!! ERROR:\033[0m\n"
}

mkdir -p selfsigned
cd selfsigned

openssl genrsa -out ca.key 1024
openssl req \
    -days 365 \
    -key ca.key \
    -new \
    -nodes \
    -out ca.pem \
    -x509 \
    -config <(
        cat <<EOF
[req]
default_md = sha256
distinguished_name = dn
prompt = no
x509_extensions = root_ca

[dn]
C = GB
CN = DappUI CA
L = London
O = DappUI
OU = DappUI Security
ST = Greater London
emailAddress = security@dappui.com

[root_ca]
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
EOF
    )

touch ca.db
echo 00 >ca.serial

mkdir -p tmp

openssl ecparam -genkey -name secp521r1 -out tls.key
openssl req \
    -days 365 \
    -key tls.key \
    -new \
    -out tls.csr \
    -config <(
        cat <<EOF
[req]
default_md = sha256
distinguished_name = dn
prompt = no
req_extensions  = server_req

[dn]
C = GB
CN = dappui.com
L = London
O = DappUI
OU = DappUI Security
ST = Greater London
emailAddress = security@dappui.com

[server_req]
basicConstraints = CA:false
keyUsage = nonRepudiation, digitalSignature, keyEncipherment, dataEncipherment, keyCertSign, cRLSign
extendedKeyUsage = serverAuth, clientAuth, timeStamping
subjectAltName = DNS:dappui.com
subjectKeyIdentifier = hash
EOF
    )

openssl ca \
    -batch \
    -days 365 \
    -in tls.csr \
    -notext \
    -out tls.cert \
    -config <(
        cat <<EOF
[ca]
default_ca = root_ca

[policy_strict]
commonName = supplied
countryName = match
emailAddress = match
organizationName = match
organizationalUnitName = match
stateOrProvinceName = match

[root_ca]
certificate = ca.pem
copy_extensions = copyall
database = ca.db
default_md = sha256
new_certs_dir = tmp
policy = policy_strict
private_key = ca.key
serial = ca.serial
EOF
    )

rm tls.csr
rm -rf tmp
rm ca.db*
rm ca.serial*
