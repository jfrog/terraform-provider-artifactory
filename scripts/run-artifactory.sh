#!/usr/bin/env sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8080:8081  releases-docker.jfrog.io/jfrog/artifactory-pro:7.24.3

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password http://localhost:8080/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""
# Use decrypted passwords
curl -u admin:password  --output /dev/null --silent --fail localhost:8080/artifactory/api/system/decrypt -X POST
