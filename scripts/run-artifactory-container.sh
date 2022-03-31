#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.27.10}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}"

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8081:8081 -p8082:8082 -p8080:8080 releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password http://localhost:8081/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""
