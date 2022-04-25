#!/usr/bin/env sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.37.14}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}"

set -euf

docker-compose --project-directory ${SCRIPT_DIR} up -d --remove-orphans

echo "Waiting for Artifactory 1 to start"
until curl -sf -u admin:password http://localhost:8081/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""

echo "Waiting for Artifactory 2 to start"
until curl -sf -u admin:password http://localhost:9081/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""

echo "Setting base URL for Artifactory 2. (Base URL for Artifactory 1 will be set by acceptance tests)"
curl -X PUT http://localhost:9081/artifactory/api/system/configuration/baseUrl -d 'http://artifactory-2:8081' -u admin:password -H "Content-type: text/plain"

# docker cp doesn't support coping files between containers so copy to local disk first
CONTAINER_ID_1=$(docker ps -q --filter ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION} --filter publish=8080)
CONTAINER_ID_2=$(docker ps -q --filter ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION} --filter publish=9080)

echo "Fetching root certificates"
docker cp ${CONTAINER_ID_1}:/opt/jfrog/artifactory/var/etc/access/keys/root.crt ${SCRIPT_DIR}/artifactory-1.crt \
  && chmod go+rw ${SCRIPT_DIR}/artifactory-1.crt
docker cp ${CONTAINER_ID_2}:/opt/jfrog/artifactory/var/etc/access/keys/root.crt ${SCRIPT_DIR}/artifactory-2.crt \
  && chmod go+rw ${SCRIPT_DIR}/artifactory-2.crt

echo "Uploading root certificates"
docker cp ${SCRIPT_DIR}/artifactory-1.crt ${CONTAINER_ID_2}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-1.crt
docker cp ${SCRIPT_DIR}/artifactory-2.crt ${CONTAINER_ID_1}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-2.crt

echo "Circle-of-Trust is setup between artifactory-1 and artifactory-2 instances"
