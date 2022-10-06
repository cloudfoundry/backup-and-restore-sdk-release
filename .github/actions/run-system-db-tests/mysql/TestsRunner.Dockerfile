FROM mysql:5.7-debian as mysql57
FROM mysql:8.0-debian as mysql80

FROM golang:1.19.1

ENV MYSQL_USERNAME= \
    MYSQL_PASSWORD= \
    MYSQL_HOSTNAME= \
    MYSQL_PORT= \
    MYSQL_CERTS_PATH=/mysql-certs

VOLUME /backup-and-restore-sdk-release

RUN mkdir -p /mysql-certs && chmod -R 777 /mysql-certs
VOLUME /mysql-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN go install github.com/onsi/ginkgo/ginkgo@latest

COPY --from=mysql57 /usr/bin/mysql /usr/local/bin/mysql57
COPY --from=mysql57 /usr/bin/mysqldump /usr/local/bin/mysqldump57
COPY --from=mysql80 /usr/bin/mysql /usr/local/bin/mysql80
COPY --from=mysql80 /usr/bin/mysqldump /usr/local/bin/mysqldump80
