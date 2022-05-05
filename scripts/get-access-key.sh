#!/usr/bin/env bash
set -x

function getAccessKey() {

echo "Generate Admin Access Key" > /dev/tty

COOKIES=$(curl -s -c - "${ARTIFACTORY_URL}/ui/api/v1/ui/auth/login?_spring_security_remember_me=false" \
              --header "accept: application/json, text/plain, */*" \
              --header "content-type: application/json;charset=UTF-8" \
              --header "x-requested-with: XMLHttpRequest" \
              -d '{"user":"admin","password":'\""${ARTIFACTORY_ADMIN_PASSWORD}"\"',"type":"login"}' | grep TOKEN)

REFRESH_TOKEN=$(echo $COOKIES | grep REFRESHTOKEN | awk '{print $7 }')
ACCESS_TOKEN=$(echo $COOKIES | grep ACCESSTOKEN | awk '{print $14 }')

ACCESS_KEY=$(curl -s -g --request GET "${ARTIFACTORY_URL}/ui/api/v1/system/security/token?services[]=all" \
                    --header "accept: application/json, text/plain, */*" \
                    --header "x-requested-with: XMLHttpRequest" \
                    --header "cookie: ACCESSTOKEN=${ACCESS_TOKEN}; REFRESHTOKEN=${REFRESH_TOKEN}")

echo "Artifactory Admin Access Key for ${ARTIFACTORY_URL}:" > /dev/tty
echo "export JFROG_ACCESS_KEY=${ACCESS_KEY}" > /dev/tty
}