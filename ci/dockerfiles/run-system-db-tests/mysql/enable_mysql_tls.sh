#!/usr/bin/env bash

set -euo pipefail

mysql_ssl_rsa_setup
cp /var/lib/mysql/ca.pem /tls-certs/
cp /var/lib/mysql/server-cert.pem /tls-certs/
cp /var/lib/mysql/server-key.pem /tls-certs/
 
mkdir -p /etc/mysql/mysql.conf.d/
cat << EOF > /etc/mysql/mysql.conf.d/ssl.cnf
[mysqld]
ssl-ca=/var/lib/mysql/ca.pem
ssl-cert=/var/lib/mysql/server-cert.pem
ssl-key=/var/lib/mysql/server-key.pem
EOF

