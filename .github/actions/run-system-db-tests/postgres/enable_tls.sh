#!/usr/bin/env bash

set -euo pipefail

# https://portal2portal.blogspot.com/2021/09/openssl-get-your-subject-right.html
# Whatever method you use to generate the certificate and key files, the Common Name
# value used for the server and client certificates/keys must each differ from the
# Common Name value used for the CA certificate.
# Otherwise, the certificate and key files will not work for servers compiled using OpenSSL.

openssl genrsa 2048 > /tls-certs/ca-key.pem
SUBJ="/C=DK/CN=CertificateAuthority/ST=SomeLand/L=SomeLocality/S=SomeProvince/O=SomeOrganization"
openssl req -new -x509 -nodes -days 3650 -key /tls-certs/ca-key.pem -subj "$SUBJ" > /tls-certs/ca-cert.pem

SUBJ="/C=IE/CN=ServerCert/ST=SomeOtherLand/L=SomeOtherLocality/S=SomeOtherProvince/O=SomeOtherOrganization"
openssl req -newkey rsa:2048 -days 3650 -nodes -keyout /tls-certs/server-key.pem -subj "$SUBJ" > /tls-certs/server-req.pem
openssl x509 -req -in /tls-certs/server-req.pem -days 3650 -CA /tls-certs/ca-cert.pem -CAkey /tls-certs/ca-key.pem -set_serial 01 > /tls-certs/server-cert.pem
openssl rsa -in /tls-certs/server-key.pem -out /tls-certs/server-key.pem

SUBJ="/C=CB/CN=ClientCert/ST=EvenOtherLand/L=EvenOtherLocality/S=EvenOtherProvince/O=EvenOtherOrganization"
openssl req -newkey rsa:2048 -days 3650 -nodes -keyout /tls-certs/client-key.pem -subj "$SUBJ" > /tls-certs/client-req.pem
openssl x509 -req -in /tls-certs/client-req.pem -days 3650 -CA /tls-certs/ca-cert.pem -CAkey /tls-certs/ca-key.pem -set_serial 01 > /tls-certs/client-cert.pem
openssl rsa -in /tls-certs/client-key.pem -out /tls-certs/client-key.pem

# https://mariadb.com/kb/en/mariadb-ssl-connection-issues/
# The core of the issue, you've used exactly the same information both for the client and the server certificate and OpenSSL doesn't like that
openssl verify -CAfile /tls-certs/ca-cert.pem /tls-certs/server-cert.pem /tls-certs/client-cert.pem

cat << EOF > /var/lib/postgresql/data/pg_hba.conf
hostssl all  all  0.0.0.0/0  md5 
EOF
