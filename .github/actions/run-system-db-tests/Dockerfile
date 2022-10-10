FROM scratch
COPY --from=dockerize-release / /

VOLUME /backup-and-restore-sdk-release

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

ENV PATH=/var/vcap/packages/golang-1-linux/bin/:${PATH}
RUN GOBIN=/usr/local/bin/ go install github.com/onsi/ginkgo/ginkgo@latest
