ARG MYSQL_IMAGE=mysql:8.0-debian@sha256:9382e4f6f7f013231049ab854de9d0810594d028f4237b67102acc7ade8e3a99
FROM ${MYSQL_IMAGE}

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN mkdir -p /etc/mysql/mysql.conf.d/ && chown mysql: /etc/mysql/mysql.conf.d/
ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

