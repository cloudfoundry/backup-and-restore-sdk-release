FROM cloudfoundry/bosh-cli

ENV BOSH_CREDS_SCRIPT ""
VOLUME /backup-and-restore-sdk-release

COPY list-bosh-releases.bash /

ENTRYPOINT /list-bosh-releases.bash
