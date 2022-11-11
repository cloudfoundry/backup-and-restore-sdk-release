#/usr/bin/env bash

# https://github.com/docker/docker/blob/master/project/PACKAGERS.md#runtime-dependencies
set -eux

apt-get update && apt-get install -y --no-install-recommends \
  btrfs-progs \
  e2fsprogs \
  iptables \
  openssl \
  ssh \
  uidmap \
  xfsprogs \
  xz-utils \
  curl \
  ca-certificates
# pigz: https://github.com/moby/moby/pull/35697 (faster gzip implementation)
# pigz \
# wget \

# set up subuid/subgid so that "--userns-remap=default" works out-of-the-box
addgroup --system dockremap
useradd --system -g dockremap dockremap
echo 'dockremap:165536:65536' >> /etc/subuid
echo 'dockremap:165536:65536' >> /etc/subgid

# https://github.com/docker/docker/tree/master/hack/dind
curl -o /usr/local/bin/dind -L "https://raw.githubusercontent.com/docker/docker/ed89041433a031cafc0a0f19cfe573c31688d377/hack/dind"
chmod +x /usr/local/bin/dind

apt-get install -y --no-install-recommends \
  apt-transport-https \
  ca-certificates \
  curl \
  gnupg-agent \
  software-properties-common \
  git-core

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository \
 "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
 $(lsb_release -cs) \
 stable"

apt-get update && apt-get install ruby bash -y --no-install-recommends
apt-get install docker-ce docker-ce-cli containerd.io -y --no-install-recommends
apt-get install jq -y --no-install-recommends
