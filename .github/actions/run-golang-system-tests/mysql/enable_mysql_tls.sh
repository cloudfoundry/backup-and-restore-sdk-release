#!/usr/bin/env bash

set -euo pipefail

mysql_ssl_rsa_setup
cp /var/lib/mysql/ca.pem /mysql-certs/
cp /var/lib/mysql/server-cert.pem /mysql-certs/
cp /var/lib/mysql/server-key.pem /mysql-certs/
 
mkdir -p /etc/mysql/mysql.conf.d/
cat << EOF > /etc/mysql/mysql.conf.d/ssl.cnf
[mysqld]
ssl-ca=/var/lib/mysql/ca.pem
ssl-cert=/var/lib/mysql/server-cert.pem
ssl-key=/var/lib/mysql/server-key.pem
EOF

