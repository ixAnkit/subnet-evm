#!/usr/bin/env bash
set -e

# Load the versions
SUBNET_EVM_PATH=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)
source "$SUBNET_EVM_PATH"/scripts/versions.sh

# Load the constants
source "$SUBNET_EVM_PATH"/scripts/constants.sh

VERSION=$METALGO_VERSION

############################
# download metalgo
# https://github.com/MetalBlockchain/metalgo/releases
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
BASEDIR=${BASE_DIR:-"/tmp/metalgo-release"}
mkdir -p ${BASEDIR}
AVAGO_DOWNLOAD_URL=https://github.com/MetalBlockchain/metalgo/releases/download/${VERSION}/metalgo-linux-${GOARCH}-${VERSION}.tar.gz
AVAGO_DOWNLOAD_PATH=${BASEDIR}/metalgo-linux-${GOARCH}-${VERSION}.tar.gz
if [[ ${GOOS} == "darwin" ]]; then
  AVAGO_DOWNLOAD_URL=https://github.com/MetalBlockchain/metalgo/releases/download/${VERSION}/metalgo-macos-${VERSION}.zip
  AVAGO_DOWNLOAD_PATH=${BASEDIR}/metalgo-macos-${VERSION}.zip
fi

METALGO_BUILD_PATH=${METALGO_BUILD_PATH:-${BASEDIR}/metalgo-${VERSION}}
mkdir -p $METALGO_BUILD_PATH

if [[ ! -f ${AVAGO_DOWNLOAD_PATH} ]]; then
  echo "downloading metalgo ${VERSION} at ${AVAGO_DOWNLOAD_URL} to ${AVAGO_DOWNLOAD_PATH}"
  curl -L ${AVAGO_DOWNLOAD_URL} -o ${AVAGO_DOWNLOAD_PATH}
fi
echo "extracting downloaded metalgo to ${METALGO_BUILD_PATH}"
if [[ ${GOOS} == "linux" ]]; then
  mkdir -p ${METALGO_BUILD_PATH} && tar xzvf ${AVAGO_DOWNLOAD_PATH} --directory ${METALGO_BUILD_PATH} --strip-components 1
elif [[ ${GOOS} == "darwin" ]]; then
  unzip ${AVAGO_DOWNLOAD_PATH} -d ${METALGO_BUILD_PATH}
  mv ${METALGO_BUILD_PATH}/build/* ${METALGO_BUILD_PATH}
  rm -rf ${METALGO_BUILD_PATH}/build/
fi

METALGO_PATH=${METALGO_BUILD_PATH}/metalgo
METALGO_PLUGIN_DIR=${METALGO_BUILD_PATH}/plugins

echo "Installed MetalGo release ${VERSION}"
echo "MetalGo Path: ${METALGO_PATH}"
echo "Plugin Dir: ${METALGO_PLUGIN_DIR}"
