#!/bin/bash

DIR="${DIR:-"$HOME/.local/bin"}"
BINARY="${BINARY:-"uboa"}"

LATEST_TAG=$(curl -s https://api.github.com/repos/oddegen/uboa/tags | jq -r '.[0].name')

ASSET_NAME="uboa-${LATEST_TAG}-linux_x86_64"

DOWNLOAD_URL="https://github.com/oddegen/uboa/releases/download/${LATEST_TAG}/${ASSET_NAME}"

curl -L -o "${ASSET_NAME}" "${DOWNLOAD_URL}"

chmod +x "${ASSET_NAME}"

mv "${ASSET_NAME}" "${DIR}/${BINARY}"

echo "Installed ${ASSET_NAME} to ${DIR}"
echo "Make sure ${DIR} is in your PATH."
