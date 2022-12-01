FROM alpine AS third-party-deps
RUN apk --update add git curl --no-cache
RUN git clone https://github.com/cloudfoundry/bosh-deployment
RUN curl -o bosh -L https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64 && chmod +x bosh

FROM ubuntu:bionic as dind-ubuntu
COPY install-dind.sh /install-dind.sh
RUN chmod +x /install-dind.sh && /install-dind.sh
VOLUME /var/lib/docker
EXPOSE 2375 2376
CMD []

FROM dind-ubuntu
COPY --from=third-party-deps bosh /usr/local/bin/bosh
COPY --from=third-party-deps bosh-deployment/ /bosh-deployment
VOLUME /backup-and-restore-sdk-release
ENTRYPOINT /backup-and-restore-sdk-release/ci/dockerfiles/bosh-in-docker/create-env.sh
