ARG MARIADB_VERSION
FROM mariadb:$MARIADB_VERSION

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN mkdir -p /etc/mysql/mariadb.conf.d/ && chown mysql: /etc/mysql/mariadb.conf.d/
ADD enable_tls.sh /docker-entrypoint-initdb.d/
