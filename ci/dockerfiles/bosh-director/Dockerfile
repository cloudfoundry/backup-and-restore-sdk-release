FROM bosh/docker-cpi:main

RUN touch /run-bosh-tests.bash           \
 && chmod +x /run-bosh-tests.bash        \
 && echo '#!/usr/bin/env bash          \n\
set -e                                 \n\
. start-bosh                           \n\
source /tmp/local-bosh/director/env    \n\
bosh --version                         \n\
bosh stemcells                         \n\
' > /run-bosh-tests.bash

ENTRYPOINT /run-bosh-tests.bash
