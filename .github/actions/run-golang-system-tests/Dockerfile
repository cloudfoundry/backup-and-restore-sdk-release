FROM mysql:8.0 as db

FROM golang:1.19.1

ENV MYSQL_USERNAME= \
    MYSQL_PASSWORD= \
    MYSQL_HOSTNAME= \
    MYSQL_PORT= \
    TEST_TLS_VERIFY_IDENTITY= \
    TEST_TLS_MUTUAL_TLS=
VOLUME /goproject

RUN go install github.com/onsi/ginkgo/ginkgo@latest

COPY --from=db /usr/bin/mysql /
COPY --from=db /usr/bin/mysqldump /
COPY task.bash /
COPY backup /
COPY restore /
ENTRYPOINT /task.bash
