FROM alpine as bosh-cli
RUN apk add curl git go --no-cache
RUN curl -o /bosh -L https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64 && chmod +x /bosh
RUN git clone https://github.com/fejnartal/bosh-cli.git /bosh-cli
WORKDIR /bosh-cli/
RUN go build -o bosh

FROM alpine
RUN apk add bash jq curl git --no-cache
COPY --from=bosh-cli /bosh-cli/bosh /usr/local/bin/bosh
