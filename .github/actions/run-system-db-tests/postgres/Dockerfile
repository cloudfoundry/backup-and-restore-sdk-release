ARG POSTGRES_VERSION
FROM postgres:$POSTGRES_VERSION

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

ADD enable_tls.sh /docker-entrypoint-initdb.d/
ADD config.sql /docker-entrypoint-initdb.d/

