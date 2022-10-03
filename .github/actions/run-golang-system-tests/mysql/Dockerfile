# Credits to: https://github.com/wixyvir/docker-mysql-tls
FROM mysql:8.0

VOLUME /mysql-certs

RUN mkdir -p /etc/mysql/mysql.conf.d/ \
 && chown mysql: /etc/mysql/mysql.conf.d/ \
 && mkdir -p /mysql-certs/ \
 && chmod 777 /mysql-certs/

ADD enable_mysql_tls.sh /docker-entrypoint-initdb.d/

