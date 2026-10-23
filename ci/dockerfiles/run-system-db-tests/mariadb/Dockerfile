ARG MARIADB_IMAGE=mariadb:10.9-jammy@sha256:56710811b0b96233ef84f4982d3337d88d3aa46a68f3fcb2853b94e1d04ea25c
FROM ${MARIADB_IMAGE}

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN mkdir -p /etc/mysql/mariadb.conf.d/ && chown mysql: /etc/mysql/mariadb.conf.d/
ADD enable_tls.sh /docker-entrypoint-initdb.d/
