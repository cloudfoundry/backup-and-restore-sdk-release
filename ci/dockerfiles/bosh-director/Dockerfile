FROM bosh/docker-cpi:main

COPY start-bosh.sh /usr/local/bin/start-bosh-debug.sh
RUN chmod +x /usr/local/bin/start-bosh-debug.sh
RUN touch /run-bosh-tests.bash           \
 && chmod +x /run-bosh-tests.bash        \
 && echo '#!/usr/bin/env bash          \n\
set -ex                                \n\
. /usr/local/bin/start-bosh-debug.sh   \n\
source /tmp/local-bosh/director/env    \n\
bosh --version                         \n\
bosh stemcells                         \n\
' > /run-bosh-tests.bash

ENTRYPOINT /run-bosh-tests.bash
