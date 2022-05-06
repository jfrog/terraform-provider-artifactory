#!/usr/bin/env bash
set -x

function getAccessKey() {
  local url=${1?You must supply the artifactory url to obtain an access key}
  echo "Generate Admin Access Key" > /dev/stderr

  local cookies
  cookies=$(curl -s -c - "${url}/ui/api/v1/ui/auth/login?_spring_security_remember_me=false" \
                --header "accept: application/json, text/plain, */*" \
                --header "content-type: application/json;charset=UTF-8" \
                --header "x-requested-with: XMLHttpRequest" \
                -d '{"user":"admin","password":"password","type":"login"}' | grep TOKEN)

  local refresh_token
  refresh_token=$(echo "${cookies}" | grep REFRESHTOKEN | awk '{print $7 }')

  local access_token
  access_token=$(echo "${cookies}" | grep ACCESSTOKEN | awk '{print $7 }')

  local access_key
  access_key=$(curl -s -g --request GET "${url}/ui/api/v1/system/security/token?services[]=all" \
                      --header "accept: application/json, text/plain, */*" \
                      --header "x-requested-with: XMLHttpRequest" \
                      --header "cookie: ACCESSTOKEN=${access_token}; REFRESHTOKEN=${refresh_token}")

  echo "export JFROG_ACCESS_KEY=${access_key}" > /dev/tty
}