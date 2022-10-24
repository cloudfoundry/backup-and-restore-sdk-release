ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN mkdir -p /etc/mysql/mysql.conf.d/ && chown mysql: /etc/mysql/mysql.conf.d/
ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

