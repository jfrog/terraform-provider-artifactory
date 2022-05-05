#!/usr/bin/env bash
set -x

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
source "${SCRIPT_DIR}/get-access-key.sh"

export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.37.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}" > /dev/stderr

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8081:8081 -p8082:8082 -p8080:8080 "releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}"

export ARTIFACTORY_URL=http://localhost:8081

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password ${ARTIFACTORY_URL}/artifactory/api/system/licenses/; do
    printf '.' > /dev/stderr
    sleep 4
done
echo ""

# with this trick you can do $(./run-artifactory-container.sh) and it will directly be setup for you
echo "export JFROG_ACCESS_KEY=$(getAccessKey "${ARTIFACTORY_URL}")"