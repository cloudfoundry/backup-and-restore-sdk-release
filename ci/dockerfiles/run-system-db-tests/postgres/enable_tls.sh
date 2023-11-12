#!/usr/bin/env bash

set -euo pipefail

ENABLE_TLS="${ENABLE_TLS:-no}"

if [[ "${ENABLE_TLS}" == "yes" || "${ENABLE_TLS}" = "mutual" ]]; then
# https://portal2portal.blogspot.com/2021/09/openssl-get-your-subject-right.html
# Whatever method you use to generate the certificate and key files, the Common Name
# value used for the server and client certificates/keys must each differ from the
# Common Name value used for the CA certificate.
# Otherwise, the certificate and key files will not work for servers compiled using OpenSSL.

openssl genrsa 2048 > /tls-certs/ca-key.pem
SUBJ="/C=DK/CN=CertificateAuthority/ST=SomeLand/L=SomeLocality/S=SomeProvince/O=SomeOrganization"
openssl req -new -x509 -nodes -days 3650 -key /tls-certs/ca-key.pem -subj "$SUBJ" > /tls-certs/ca-cert.pem

# Notice that we are using `system-db-postgres-backing-db`
# which is the same name we give to the Postgres backing db service
# this is needed for the plsql Mutual TLS to work correctly
SUBJ="/C=IE/CN=system-db-postgres-backing-db/ST=SomeOtherLand/L=SomeOtherLocality/S=SomeOtherProvince/O=SomeOtherOrganization"
openssl req -newkey rsa:2048 -days 3650 -nodes -keyout /tls-certs/server-key.pem -subj "$SUBJ" > /tls-certs/server-req.pem
openssl x509 -req -in /tls-certs/server-req.pem -days 3650 -CA /tls-certs/ca-cert.pem -CAkey /tls-certs/ca-key.pem -set_serial 01 > /tls-certs/server-cert.pem
openssl rsa -in /tls-certs/server-key.pem -out /tls-certs/server-key.pem

# Notice that we are using `postgres`
# whish is the same name we give to the POSTGRES_USERNAME variable
# this is needed for the plsql Mutual TLS to work correctly in `verify-full` mode
# https://postgrespro.com/list/thread-id/2372010
SUBJ="/C=CB/CN=postgres/ST=EvenOtherLand/L=EvenOtherLocality/S=EvenOtherProvince/O=EvenOtherOrganization"
openssl req -newkey rsa:2048 -days 3650 -nodes -keyout /tls-certs/client-key.pem -subj "$SUBJ" > /tls-certs/client-req.pem
openssl x509 -req -in /tls-certs/client-req.pem -days 3650 -CA /tls-certs/ca-cert.pem -CAkey /tls-certs/ca-key.pem -set_serial 01 > /tls-certs/client-cert.pem
openssl rsa -in /tls-certs/client-key.pem -out /tls-certs/client-key.pem

# https://mariadb.com/kb/en/mariadb-ssl-connection-issues/
# The core of the issue, you've used exactly the same information both for the client and the server certificate and OpenSSL doesn't like that
openssl verify -CAfile /tls-certs/ca-cert.pem /tls-certs/server-cert.pem /tls-certs/client-cert.pem

cat << 'EOF' > /docker-entrypoint-initdb.d/enable-tls.sql
ALTER SYSTEM SET ssl_ca_file TO '/tls-certs/ca-cert.pem';
ALTER SYSTEM SET ssl_cert_file TO '/tls-certs/server-cert.pem';
ALTER SYSTEM SET ssl_key_file TO '/tls-certs/server-key.pem';
ALTER SYSTEM SET ssl TO 'ON';
EOF
fi

if [[ "${ENABLE_TLS}" == "yes" ]]; then
  cat << 'EOF' > /var/lib/postgresql/data/pg_hba.conf
hostssl all  all  0.0.0.0/0  md5 
EOF
  elif [[ "${ENABLE_TLS}" == "mutual" ]]; then

      if postgres --version | grep -o "15.5"; then
cat << 'EOF' > /var/lib/postgresql/data/pg_hba.conf
hostssl all  all  0.0.0.0/0  md5 clientcert=verify-full
EOF
      else
cat << 'EOF' > /var/lib/postgresql/data/pg_hba.conf
hostssl all  all  0.0.0.0/0  md5 clientcert=1
EOF
      fi

fi
