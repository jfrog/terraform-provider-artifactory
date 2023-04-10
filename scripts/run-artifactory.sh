#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
source "${SCRIPT_DIR}/get-access-key.sh"
source "${SCRIPT_DIR}/wait-for-rt.sh"
export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.55.8}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}"

set -euf

docker-compose --project-directory "${SCRIPT_DIR}" up -d --remove-orphans

ARTIFACTORY_URL_1=http://localhost:8081
ARTIFACTORY_UI_URL_1=http://localhost:8082
ARTIFACTORY_URL_2=http://localhost:9081
ARTIFACTORY_UI_URL_2=http://localhost:9082

echo "Waiting for Artifactory 1 to start"
waitForArtifactory "${ARTIFACTORY_URL_1}" "${ARTIFACTORY_UI_URL_1}"

echo "Waiting for Artifactory 2 to start"
waitForArtifactory "${ARTIFACTORY_URL_2}" "${ARTIFACTORY_UI_URL_2}"

echo "Setting base URL for Artifactory 2. (Base URL for Artifactory 1 will be set by acceptance tests)"
curl -X PUT "${ARTIFACTORY_URL_2}/artifactory/api/system/configuration/baseUrl" -d 'http://artifactory-2:8081' -u admin:password -H "Content-type: text/plain"

# docker cp doesn't support copying files between containers so copy to local disk first
CONTAINER_ID_1=$(docker ps -q --filter "ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}" --filter publish=8080)
CONTAINER_ID_2=$(docker ps -q --filter "ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}" --filter publish=9080)

echo "Fetching root certificates"
docker cp "${CONTAINER_ID_1}":/opt/jfrog/artifactory/var/etc/access/keys/root.crt "${SCRIPT_DIR}/artifactory-1.crt" \
  && chmod go+rw "${SCRIPT_DIR}"/artifactory-1.crt
docker cp "${CONTAINER_ID_2}":/opt/jfrog/artifactory/var/etc/access/keys/root.crt "${SCRIPT_DIR}/artifactory-2.crt" \
  && chmod go+rw "${SCRIPT_DIR}"/artifactory-2.crt

echo "Uploading root certificates"
docker cp "${SCRIPT_DIR}/artifactory-1.crt" "${CONTAINER_ID_2}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-1.crt"
docker cp "${SCRIPT_DIR}/artifactory-2.crt" "${CONTAINER_ID_1}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-2.crt"

echo "Circle-of-Trust is setup between artifactory-1 and artifactory-2 instances"

echo "Generate Admin Access Keys for both instances"

echo "export JFROG_ACCESS_TOKEN=$(getAccessKey ${ARTIFACTORY_UI_URL_1})"

# to be able to run federated repo tests add ARTIFACTORY_URL_2=http://host.docker.internal:9081 or ARTIFACTORY_URL_2=http://artifactory-2:9081 variable
# see https://github.com/jfrog/terraform-provider-artifactory/wiki/Testing for the details.