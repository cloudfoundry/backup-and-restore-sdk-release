#!/usr/bin/env bash

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
  gnupg \
  software-properties-common \
  ruby \
  git

# set up subuid/subgid so that "--userns-remap=default" works out-of-the-box
addgroup --system dockremap
useradd --system -g dockremap dockremap
echo 'dockremap:165536:65536' >> /etc/subuid
echo 'dockremap:165536:65536' >> /etc/subgid

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list

apt-get update && apt-get install -y --no-install-recommends docker-ce docker-ce-cli containerd.io

# Docker-in-Docker requires iptables-legacy; nftables backend breaks nested networking
update-alternatives --set iptables /usr/sbin/iptables-legacy
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
