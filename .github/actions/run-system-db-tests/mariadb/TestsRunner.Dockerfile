FROM mysql:5.7-debian as mysql57

FROM golang:1.19.1

ENV MYSQL_USERNAME= \
    MYSQL_PASSWORD= \
    MYSQL_HOSTNAME= \
    MYSQL_PORT= \
    MYSQL_CERTS_PATH=/tls-certs

VOLUME /backup-and-restore-sdk-release

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

RUN apt-get update && apt-get install mariadb-client -y --no-install-recommends && apt-get clean

RUN go install github.com/onsi/ginkgo/ginkgo@latest

COPY --from=mysql57 /usr/bin/mysql /usr/local/bin/mysql57
COPY --from=mysql57 /usr/bin/mysqldump /usr/local/bin/mysqldump57

