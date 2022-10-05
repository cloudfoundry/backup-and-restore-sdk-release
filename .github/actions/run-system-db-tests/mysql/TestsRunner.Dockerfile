ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION as db

FROM golang:1.19.1

ENV MYSQL_USERNAME= \
    MYSQL_PASSWORD= \
    MYSQL_HOSTNAME= \
    MYSQL_PORT= \
    MYSQL_CERTS_PATH=/mysql-certs

VOLUME /backup-and-restore-sdk-release
VOLUME /mysql-certs

RUN go install github.com/onsi/ginkgo/ginkgo@latest

COPY --from=db /usr/bin/mysql /usr/local/bin/
COPY --from=db /usr/bin/mysqldump /usr/local/bin/
