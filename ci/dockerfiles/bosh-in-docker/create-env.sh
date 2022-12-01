#!/usr/bin/env bash

set -euo pipefail

# To run bosh CLI using docker cpi:
# https://github.com/cloudfoundry-attic/bosh-lite/issues/439#issuecomment-348329967
dockerd > /dev/null 2>&1 &

export BOSH_LOG_LEVEL=none
bosh -n --tty create-env /bosh-deployment/bosh.yml \
  -o /bosh-deployment/docker/cpi.yml \
  -o /bosh-deployment/docker/unix-sock.yml \
  -o /bosh-deployment/jumpbox-user.yml \
  --state=/workspace/state.json              \
  --vars-store /workspace/creds.yml          \
  -v director_name=docker \
  -v internal_cidr=10.245.0.0/16 \
  -v internal_gw=10.245.0.1 \
  -v internal_ip=10.245.0.10 \
  -v docker_host=unix:///var/run/docker.sock \
  -v network=net3
##


# Docker CPI - Cannot upload stemcell due to "Cannot connect to the Docker daemon... Is the docker daemon running?"
# https://github.com/cloudfoundry/bosh-deployment/issues/94
chmod 777 /var/run/docker.sock

ssh-keygen -t rsa -q -f /shared-creds/id_rsa -N ""
mkdir -p ~/.ssh && cat /shared-creds/id_rsa.pub | cat > ~/.ssh/authorized_keys
/etc/init.d/ssh start

cat << EOF > /shared-creds/bosh-creds.bash
export BOSH_CLIENT_SECRET='$(bosh int /workspace/creds.yml --path /admin_password)'
export BOSH_CA_CERT='$(bosh int /workspace/creds.yml --path /director_ssl/ca)'
export BOSH_CLIENT=admin
export BOSH_ENVIRONMENT=https://10.245.0.10:25555
export BOSH_ALL_PROXY=ssh+sock5://root@bosh-in-docker:22?private-key=/shared-creds/id_rsa
export BOSH_GW_USER=root
export BOSH_GW_HOST=bosh-in-docker
export BOSH_GW_PRIVATE_KEY='$(cat /shared-creds/id_rsa)'
EOF

source /shared-creds/bosh-creds.bash
bosh -n --tty update-cloud-config /bosh-deployment/docker/cloud-config.yml -v network=net3

while true; do sleep 30; done;
