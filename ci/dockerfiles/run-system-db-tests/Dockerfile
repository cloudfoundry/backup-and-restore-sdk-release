FROM bosh/cli

RUN mkdir -p /tls-certs && chmod -R 777 /tls-certs
VOLUME /tls-certs
VOLUME /shared-creds
# https://boxboat.com/2017/01/23/volumes-and-dockerfiles-dont-mix/

# When BOSH_ALL_PROXY is not empty Bosh invokes "nc -x" command
# thus we need to install netcat-openbsd variant (netcat doesn't work).
RUN apt-get update && apt-get install -y --no-install-recommends netcat-openbsd

COPY ./ /backup-and-restore-sdk-release
