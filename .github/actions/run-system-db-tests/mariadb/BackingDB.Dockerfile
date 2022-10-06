ARG MARIADB_VERSION
FROM mariadb:$MARIADB_VERSION

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
# Commented out until we can determine why specifying the
# volume in the Dockerfile causes some permission issues
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/
# VOLUME /tls-certs

RUN mkdir -p /etc/mysql/mariadb.conf.d/ && chown mysql: /etc/mysql/mariadb.conf.d/
ADD enable_tls.sh /docker-entrypoint-initdb.d/
