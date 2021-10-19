# Artifactory Remote Repository Resource

Provides an Artifactory remote `docker` repository resource. This provides docker specific fields and is the only way to
get them


## Example Usage
Includes only new and relevant fields
```hcl
# Create a new Artifactory remote docker repository called my-remote-docker
resource "artifactory_remote_docker_repository" "my-remote-docker" {
  key                            = "my-remote-docker"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/hub.docker.io/**", "**/bintray.jfrog.io/**"]
  enable_token_authentication    = true
  url                            = "https://hub.docker.io/"
  block_pushing_schema1          = true
}
```

## Argument Reference

Arguments have a one to one mapping with
the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are
supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `block_pushing_schema1` - (Optional) When set, Artifactory will block the pulling of Docker images with manifest v2
  schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1
  that exist in the cache.
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS
