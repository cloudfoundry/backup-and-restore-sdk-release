ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION

VOLUME /mysql-certs

# Credits to: https://github.com/wixyvir/docker-mysql-tls
RUN mkdir -p /etc/mysql/mysql.conf.d/ \
 && chown mysql: /etc/mysql/mysql.conf.d/ \
 && mkdir -p /mysql-certs/ \
 && chmod 777 /mysql-certs/

ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

