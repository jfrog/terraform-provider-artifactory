#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
source "${SCRIPT_DIR}/get-access-key.sh"
source "${SCRIPT_DIR}/wait-for-rt.sh"

export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.84.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}" > /dev/stderr

set -euf

rm -rf ${SCRIPT_DIR}/artifactory/

mkdir -p ${SCRIPT_DIR}/artifactory/extra_conf
mkdir -p ${SCRIPT_DIR}/artifactory/var/etc/access

mkdir -p ${SCRIPT_DIR}/artifactory/var/etc/access
sudo chown -R 1030:1030 ${SCRIPT_DIR}/artifactory/var

cp ${SCRIPT_DIR}/artifactory.lic ${SCRIPT_DIR}/artifactory/extra_conf
cp ${SCRIPT_DIR}/system.yaml ${SCRIPT_DIR}/artifactory/var/etc/
cp ${SCRIPT_DIR}/access.config.patch.yml ${SCRIPT_DIR}/artifactory/var/etc/access

docker run -i --name artifactory-1 -d --rm \
  -e JF_FRONTEND_FEATURETOGGLER_ACCESSINTEGRATION=true \
  -e VAULT_ADDR \
  -e VAULT_TOKEN \
  -e VAULT_ROLE_ID \
  -e VAULT_SECRET_ID \
  -e VAULT_PATH \
  -v ${SCRIPT_DIR}/artifactory/extra_conf:/artifactory_extra_conf \
  -v ${SCRIPT_DIR}/artifactory/var:/var/opt/jfrog/artifactory \
  -p 8081:8081 -p 8082:8082 \
  releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}

export ARTIFACTORY_URL=http://localhost:8081
export ARTIFACTORY_UI_URL=https://localhost:8082

# Wait for Artifactory to start
waitForArtifactory "${ARTIFACTORY_URL}" "${ARTIFACTORY_UI_URL}"

# With this trick you can do $(./run-artifactory-container.sh) and it will directly be setup for you without the terminal output
echo "export JFROG_ACCESS_TOKEN=$(getAccessKey "${ARTIFACTORY_UI_URL}")"
