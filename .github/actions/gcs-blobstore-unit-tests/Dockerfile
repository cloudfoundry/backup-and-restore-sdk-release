FROM golang

ENV GOPATH /go
ENV PATH /go/bin:/usr/local/go/bin:$PATH
RUN go install github.com/onsi/ginkgo/ginkgo@latest

VOLUME /backup-and-restore-sdk-release

COPY run-gcs-blobstore-tests.bash /

ENTRYPOINT /run-gcs-blobstore-tests.bash
