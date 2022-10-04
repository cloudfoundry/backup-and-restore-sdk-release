ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

RUN mkdir -p /mysql-certs && chown mysql:mysql /mysql-certs
VOLUME /mysql-certs

RUN mkdir -p /etc/mysql/mysql.conf.d/ && chown mysql: /etc/mysql/mysql.conf.d/
ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

