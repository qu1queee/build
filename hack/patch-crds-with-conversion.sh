# Copyright The Shipwright Contributors
#
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

ORG=geofffranks
REPO=spruce
VERSION=v1.30.2

SYSTEM_UNAME="$(uname | tr '[:upper:]' '[:lower:]')"
SYSTEM_ARCH="$(uname -m | sed 's/x86_64/amd64/')"
TARGET_DIR=/tmp/

echo "[INFO] Retrieving spruce binary release location"
DOWNLOAD_URI="$(curl --silent --location "https://api.github.com/repos/${ORG}/${REPO}/releases/tags/${VERSION}" | jq --raw-output ".assets[] | select( (.name | contains(\"${SYSTEM_UNAME}\")) and (.name | contains(\"${SYSTEM_ARCH}\")) and (.name | contains(\"sha1\") | not) ) | .browser_download_url")"
if [[ -z ${DOWNLOAD_URI} ]]; then
  echo -e "Unsupported operating system or machine type"
  exit 1
fi


echo "[INFO] Downloading spruce binary with version ${VERSION}"
if curl --progress-bar --location "${DOWNLOAD_URI}" --output "${TARGET_DIR}/spruce"; then
  chmod a+rx "${TARGET_DIR}/spruce"
fi


echo "[INFO] Going to patch the Build CRD"
$TARGET_DIR/spruce merge $DIR/hack/customization/conversion_webhook_block.yaml $DIR/deploy/crds/shipwright.io_builds.yaml > /tmp/shipwright.io_builds.yaml
mv /tmp/shipwright.io_builds.yaml "${DIR}"/deploy/crds/shipwright.io_builds.yaml
echo "[INFO] Build CRD successfully patched"