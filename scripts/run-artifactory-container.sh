#!/usr/bin/env bash
set -x

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.37.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}"

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8081:8081 -p8082:8082 -p8080:8080 releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}

ARTIFACTORY_URL=http://localhost:8081

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password ${ARTIFACTORY_URL}/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""

echo "Generate Admin Access Key"

COOKIES=$(curl -c - "${ARTIFACTORY_URL}/ui/api/v1/ui/auth/login?_spring_security_remember_me=false" \
              --header "accept: application/json, text/plain, */*" \
              --header "content-type: application/json;charset=UTF-8" \
              --header "x-requested-with: XMLHttpRequest" \
              -d '{"user":"admin","password":"Password1!","type":"login"}' | grep TOKEN)

REFRESH_TOKEN=$(echo $COOKIES | grep REFRESHTOKEN | awk '{print $7 }')
ACCESS_TOKEN=$(echo $COOKIES | grep ACCESSTOKEN | awk '{print $14 }')

ACCESS_KEY=$(curl -g --request GET "${ARTIFACTORY_URL}/ui/api/v1/system/security/token?services[]=all" \
                    --header "accept: application/json, text/plain, */*" \
                    --header "x-requested-with: XMLHttpRequest" \
                    --header "cookie: ACCESSTOKEN=${ACCESS_TOKEN}; REFRESHTOKEN=${REFRESH_TOKEN}")
echo "Artifactory Admin Access Key: ${ACCESS_KEY}"