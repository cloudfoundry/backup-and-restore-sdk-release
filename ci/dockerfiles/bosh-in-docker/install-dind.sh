#/usr/bin/env bash

# https://github.com/docker/docker/blob/master/project/PACKAGERS.md#runtime-dependencies
set -eu

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
  ca-certificates \
  apt-transport-https \
  gnupg-agent \
  software-properties-common \
  ruby

# set up subuid/subgid so that "--userns-remap=default" works out-of-the-box
addgroup --system dockremap
useradd --system -g dockremap dockremap
echo 'dockremap:165536:65536' >> /etc/subuid
echo 'dockremap:165536:65536' >> /etc/subgid

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository \
 "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
 $(lsb_release -cs) \
 stable"

apt-get install docker-ce docker-ce-cli containerd.io -y --no-install-recommends
