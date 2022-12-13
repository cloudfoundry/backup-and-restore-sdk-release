FROM cloudfoundry/bosh-cli

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
VOLUME /shared-creds
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

# Required to run the tests
RUN go install github.com/onsi/ginkgo/ginkgo@latest

# Tests assume the bosh binary is called bosh-cli
RUN ln -s /usr/bin/bosh /usr/bin/bosh-cli

# When BOSH_ALL_PROXY is not empty Bosh invokes "nc -x" command
# thus we need to install netcat-openbsd variant (netcat doesn't work).
RUN apt-get install -y --no-install-recommends netcat-openbsd

COPY ./ /backup-and-restore-sdk-release
