ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

VOLUME /mysql-certs
RUN mkdir -p /mysql-certs && chmod -R 777 /mysql-certs

RUN mkdir -p /etc/mysql/mysql.conf.d/ && chown mysql: /etc/mysql/mysql.conf.d/
ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

