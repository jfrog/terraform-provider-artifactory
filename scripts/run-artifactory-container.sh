#!/usr/bin/env bash
set -x

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
. ${SCRIPT_DIR}/get-access-key.sh

export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.37.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}"
ARTIFACTORY_ADMIN_PASSWORD="password"

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8081:8081 -p8082:8082 -p8080:8080 "releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}"

export ARTIFACTORY_URL=http://localhost:8081

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password ${ARTIFACTORY_URL}/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""

# by running this script - $(./run-artifactory-container.sh) all needed variables to run the provider will be setup
getAccessKey > /dev/null 2>&1