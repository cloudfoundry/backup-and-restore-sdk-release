FROM alpine@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS third-party-deps
RUN apk --update add git curl --no-cache
RUN git clone https://github.com/cloudfoundry/bosh-deployment \
    && git -C bosh-deployment checkout f3b0d9c5d9b040f1eeeb2554a5c0855375a4c399
RUN curl --fail -o bosh -L https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64 \
    && printf '%s  %s\n' '808ed5b94c45ac1f5a351ace62f2a93adb6fcf2f689368cbba12b37343f1121e' bosh | sha256sum -c - \
    && chmod +x bosh

FROM ubuntu:bionic@sha256:152dc042452c496007f07ca9127571cb9c29697f42acbfad72324b2bb2e43c98 as dind-ubuntu
COPY ./ci/dockerfiles/bosh-in-docker/install-dind.sh /install-dind.sh
RUN chmod +x /install-dind.sh && /install-dind.sh
VOLUME /var/lib/docker
EXPOSE 2375 2376
CMD []

FROM dind-ubuntu
COPY --from=third-party-deps bosh /usr/local/bin/bosh
COPY --from=third-party-deps bosh-deployment/ /bosh-deployment
COPY ./ /backup-and-restore-sdk-release
ENTRYPOINT /backup-and-restore-sdk-release-ci/ci/dockerfiles/bosh-in-docker/create-env.sh
