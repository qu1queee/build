#!/bin/bash

# Copyright The Shipwright Contributors
# 
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "Logging into container registry $IMAGE_HOST"
echo "$REGISTRY_PASSWORD" | ko login -u "$REGISTRY_USERNAME" --password-stdin "$IMAGE_HOST"

echo "Building container image"

# Using defaults, this pushes to:
# quay.io/shipwright/shipwright-operator:latest
KO_DOCKER_REPO="quay.io/encaladaenrique/eeeoo" GOFLAGS="${GO_FLAGS}" ko resolve -t "$TAG" --bare -R -f deploy/ > release.yaml
KO_DOCKER_REPO="quay.io/encaladaenrique/eeeoo" GOFLAGS="${GO_FLAGS} -tags=pprof_enabled" ko resolve -t "$TAG-debug" --bare -R -f deploy/ > release-debug.yaml
