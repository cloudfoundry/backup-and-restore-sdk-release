ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

RUN mkdir -p /mysql-certs && chmod -R 777 /mysql-certs
# Commented out until we can determine why specifying the
# volume in the Dockerfile causes some permission issues
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/
# VOLUME /mysql-certs

RUN mkdir -p /etc/mysql/mysql.conf.d/ && chown mysql: /etc/mysql/mysql.conf.d/
ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

