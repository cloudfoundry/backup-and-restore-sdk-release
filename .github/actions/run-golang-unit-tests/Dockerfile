FROM golang:1.19.1

VOLUME /goproject

RUN go install github.com/onsi/ginkgo/ginkgo@latest

ENTRYPOINT /goproject/scripts/run-unit-tests.bash
