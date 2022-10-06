ARG MARIADB_VERSION
FROM mariadb:$MARIADB_VERSION

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
# Commented out until we can determine why specifying the
# volume in the Dockerfile causes some permission issues
# VOLUME /tls-certs

RUN mkdir -p /etc/mysql/mariadb.conf.d/ && chown mysql: /etc/mysql/mariadb.conf.d/
ADD enable_tls.sh /docker-entrypoint-initdb.d/
