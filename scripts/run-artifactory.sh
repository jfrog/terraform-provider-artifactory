#!/usr/bin/env sh

set -euf

docker run -i -t -d --rm -v "${PWD}/scripts/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" -p8080:8081 --name artifactory docker.bintray.io/jfrog/artifactory-pro:6.6.5

echo "Waiting for Artifactory to start"
until curl --output /dev/null --silent --head --fail http://localhost:8080/artifactory/webapp/#/login; do
    echo '.'
    sleep 4
done

# Use decrypted passwords
curl -u admin:password  --output /dev/null --silent --fail localhost:8080/artifactory/api/system/decrypt -X POST
