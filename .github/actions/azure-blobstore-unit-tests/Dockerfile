FROM golang:1.19.1

VOLUME /backup-and-restore-sdk-release

RUN go install github.com/onsi/ginkgo/ginkgo@latest
COPY run-azure-blobstore-tests.bash /

ENTRYPOINT /run-azure-blobstore-tests.bash
