# artifactory_remote_repository

Provides an Artifactory remote repository resource. This can be used to create and manage Artifactory remote repositories.

## Example Usage

```hcl
# Create a new Artifactory remote repository called my-remote
resource "artifactory_remote_repository" "my-remote" {
  key             = "my-remote"
  package_type    = "npm"
  url             = "https://registry.npmjs.org/"
  repo_layout_ref = "npm-default"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `package_type` - (Required)
* `url` - (Required)
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `handle_releases` - (Optional)
* `handle_snapshots` - (Optional)
* `max_unique_snapshots` - (Optional)
* `suppress_pom_consistency_checks` - (Optional)
* `username` - (Optional)
* `password` - (Optional) Requires password encryption to be turned off `POST /api/system/decrypt`
* `proxy` - (Optional)
* `hard_fail` - (Optional)
* `offline` - (Optional)
* `blacked_out` - (Optional)
* `store_artifacts_locally` - (Optional)
* `socket_timeout_millis` - (Optional)
* `local_address` - (Optional)
* `retrieval_cache_period_seconds` - (Optional)
* `missed_cache_period_seconds` - (Optional)
* `unused_artifacts_cleanup_period_hours` - (Optional)
* `fetch_jars_eagerly` - (Optional)
* `fetch_sources_eagerly` - (Optional)
* `share_configuration` - (Optional)
* `synchronize_properties` - (Optional)
* `block_mismatching_mime_types` - (Optional)
* `property_sets` - (Optional)
* `allow_any_host_auth` - (Optional)
* `enable_cookie_management` - (Optional)
* `client_tls_certificate` - (Optional)
* `pypi_registry_url` - (Optional)
* `bypass_head_requests` - (Optional)
* `enable_token_authentication` - (Optional)


## Import

Remote repositories can be imported using their name, e.g.

```
$ terraform import artifactory_remote_repository.my-remote my-remote
```