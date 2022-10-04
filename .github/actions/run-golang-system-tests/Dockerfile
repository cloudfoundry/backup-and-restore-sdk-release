ARG MYSQL_VERSION
FROM mysql:$MYSQL_VERSION as db

FROM golang:1.19.1

ENV MYSQL_USERNAME= \
    MYSQL_PASSWORD= \
    MYSQL_HOSTNAME= \
    MYSQL_PORT= \
    TEST_TLS= \
    TEST_TLS_VERIFY_IDENTITY= \
    TEST_TLS_MUTUAL_TLS= \
    MYSQL_CA_CERT_PATH= \
    MYSQL_CLIENT_CERT_PATH= \
    MYSQL_CLIENT_KEY_PATH=

VOLUME /goproject
VOLUME /mysql-certs
RUN mkdir -p /mysql-certs && chmod -R 777 /mysql-certs

RUN go install github.com/onsi/ginkgo/ginkgo@latest

COPY --from=db /usr/bin/mysql /
COPY --from=db /usr/bin/mysqldump /
COPY task.bash /
COPY backup /
COPY restore /
ENTRYPOINT /task.bash
