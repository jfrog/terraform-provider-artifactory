# Artifactory Remote Repository Resource

Provides an Artifactory remote `docker` repository resource. This provides docker specific fields and is the only way to get them

### Passwords
Passwords can only be used when encryption is turned off (https://www.jfrog.com/confluence/display/RTF/Artifactory+Key+Encryption). 
Since only the artifactory server can decrypt them it is impossible for terraform to diff changes correctly.

To get full management, passwords can be decrypted globally using `POST /api/system/decrypt`. If this is not possible, 
the password diff can be disabled per resource with-- noting that this will require resources to be tainted for an update:
```hcl
lifecycle {
    ignore_changes = ["password"]
}
``` 

## Example Usage

```hcl
# Create a new Artifactory remote docker repository called my-remote-docker
resource "artifactory_remote_docker_repository" "my-remote-docker" {
  key                                     = "my-remote-docker"
  package_type                            = "docker"
  external_dependencies_patterns          = ["**/hub.docker.io/**","**/bintray.jfrog.io/**"]
  hard_fail                               = true
  allow_any_host_auth                     = true
  external_dependencies_enabled           = true
  socket_timeout_millis                   = 25000
  retrieval_cache_period_seconds          = 70
  enable_token_authentication             = true
  property_sets                           = ["artifactory"]
  proxy                                   = ""
  store_artifacts_locally                 = true
  unused_artifacts_cleanup_period_hours   = 96
  username                                = "user"
  content_synchronisation                  {
    enabled = false
  }
  missed_cache_period_seconds             = 2500
  excludes_pattern                        = ""
  url                                     = "https://hub.docker.io/"
  share_configuration                     = true
  unused_artifacts_cleanup_period_enabled = true
  client_tls_certificate                  = ""
  xray_index                              = true
  block_mismatching_mime_types            = true
  offline                                 = true
  local_address                           = ""
  repo_layout_ref                         = "docker-default"
  notes                                   = "notes"
  enable_cookie_management                = true
  synchronize_properties                  = true
  assumed_offline_period_secs             = 96
  block_pushing_schema1                   = true
  blacked_out                             = true
  bypass_head_requests                    = true
  includes_pattern                        = "**/*"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `package_type` - (Required) - self-explanatory
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `allow_any_host_auth` - (Optional) Also known as 'Lenient Host Authentication', Allow credentials of this repository to be used on requests redirected to any other host.
* `assumed_offline_period_secs` - (Optional)
* `blacked_out` - (Optional) (A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.
* `block_mismatching_mime_types` - (Optional) Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.
* `block_pushing_schema` - (Optional) When set, Artifactory will block the pulling of Docker images with manifest v2 schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1 that exist in the cache.
* `bypass_head_requests` - (Optional)
* `client_tls_certificate` - (Optional)
* `content_synchronisation` - (Optional)
* `description` - (Optional)
* `enable_cookie_management` - (Optional) Enables cookie management if the remote repository uses cookies to manage client state.
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `enabled` - (Optional)
* `excludes_pattern` - (Optional)
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will 
  follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
  By default, this is set to '**', which means that remote modules may be downloaded from any external VCS source.
* `hard_fail` - (Optional)
* `includes_pattern` - (Optional)
* `local_address` - (Optional)
* `missed_cache_period_seconds` - (Optional)
* `notes` - (Optional)
* `offline` - (Optional) If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.
* `password` - (Optional)
* `propagate_query_params` - (Optional)
* `property_sets` - (Optional)
* `proxy` - (Optional)
* `repo_layout_ref` - (Optional)
* `retrieval_cache_period_seconds` - (Optional)
* `share_configuration` - (Optional)
* `socket_timeout_millis` - (Optional)
* `store_artifacts_locally` - (Optional) When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming it directly to satellite pass-though Artifactory servers.
* `synchronize_properties` - (Optional) When set, remote artifacts are fetched along with their properties.
* `unused_artifacts_cleanup_period_enabled` - (Optional)
* `unused_artifacts_cleanup_period_hours` - (Optional)
* `username` - (Optional)
* `xray_index` - (Optional)


## Import

Remote repositories can be imported using their name, e.g.

```
$ terraform import artifactory_remote_repository.my-remote my-remote
```
