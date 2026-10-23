ARG POSTGRES_IMAGE=postgres:15-bullseye@sha256:4ce70bcb2c05817a09b9919fc440878cf893cfb5623cd05023778a67644ea81d
FROM ${POSTGRES_IMAGE}

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

ADD enable_tls.sh /docker-entrypoint-initdb.d/
RUN touch /docker-entrypoint-initdb.d/enable-tls.sql && chmod 777 /docker-entrypoint-initdb.d/enable-tls.sql
