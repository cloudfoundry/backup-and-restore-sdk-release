FROM alpine as bosh-cli
RUN apk add curl --no-cache
RUN curl -o /bosh -L https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64 && chmod +x /bosh

FROM alpine
RUN apk add bash jq curl --no-cache
COPY --from=bosh-cli /bosh /usr/local/bin/bosh
