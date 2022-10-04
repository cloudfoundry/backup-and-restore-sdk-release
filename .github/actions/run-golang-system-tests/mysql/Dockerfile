ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

VOLUME /mysql-certs

ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

