function waitForArtifactory() {
  local url=${1?You must supply the artifactory url}
  local url_ui=${2?You must supply the artifactory UI url}
  echo "### Wait for Artifactory to start at ${url} ###" > /dev/stderr

  until $(curl -sf -o /dev/null -m 5 ${url}/artifactory/api/system/ping/); do
      printf '.'
      sleep 5
  done
  echo ""

  echo "### Waiting for Artifactory UI to start at ${url_ui} ###"
  until $(curl -sf -o /dev/null -m 5 ${url_ui}/ui/login/); do
      printf '.'
      sleep 5
  done
  echo ""
}