FROM alpine AS third-party-deps
RUN apk --update add git curl --no-cache
RUN curl -o bosh -L https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64
RUN chmod +x bosh

RUN git clone https://github.com/cloudfoundry/cf-deployment
RUN git clone https://github.com/cloudfoundry/bosh-deployment

FROM ubuntu:bionic as dind-ubuntu
COPY install-dind.sh /install-dind.sh
RUN chmod +x /install-dind.sh
RUN /install-dind.sh

VOLUME /var/lib/docker
EXPOSE 2375 2376
CMD []

FROM dind-ubuntu
COPY --from=third-party-deps bosh /usr/local/bin/bosh
COPY --from=third-party-deps bosh-deployment/ /bosh-deployment
COPY --from=third-party-deps cf-deployment/ /cf-deployment
COPY create-env.sh /create-env.sh
RUN chmod +x /create-env.sh
ENTRYPOINT /create-env.sh
