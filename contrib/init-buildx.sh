#!/usr/bin/env bash

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

export DOCKER_CLI_EXPERIMENTAL=enabled

if ! docker buildx 2>&1 >/dev/null; then
  echo "buildx not available. Docker 19.03 or higher is required with experimental features enabled"
  exit 1
fi

# We can skip setup if the current builder already has multi-arch
# AND if it isn't the docker driver, which doesn't work
current_builder="$(docker buildx inspect)"
# linux/amd64, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/arm/v7, linux/arm/v6
if ! grep -q "^Driver: docker$"  <<<"${current_builder}" && \
     grep -q "linux/amd64" <<<"${current_builder}" && \
     grep -q "linux/arm"   <<<"${current_builder}" && \
     grep -q "linux/arm64" <<<"${current_builder}"; then
  exit 0
fi

# Ensure qemu is in binfmt_misc
# Docker desktop already has these in versions recent enough to have buildx
# We only need to do this setup on linux hosts
if [ "$(uname)" == 'Linux' ]; then
  # NOTE: this is pinned to a digest for a reason!
  docker run --rm --privileged multiarch/qemu-user-static@sha256:28ebe2e48220ae8fd5d04bb2c847293b24d7fbfad84f0b970246e0a4efd48ad6 --reset -p yes
fi

# Ensure we use a builder that can leverage it (the default on linux will not)
docker buildx rm openfaas || true
docker buildx create --use --name=openfaas
