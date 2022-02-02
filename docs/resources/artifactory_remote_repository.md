# Artifactory Remote Repository Resource

Provides an Artifactory remote repository resource. This can be used to create and manage Artifactory remote repositories.

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
* `feed_context_path` - (Optional, Nuget repos only)
* `download_context_path` - (Optional, Nuget repos only)
* `v3_feed_url` - (Optional, Nuget repos only)
* `force_nuget_authentication` - (Optional, Nuget repos only) 
* `nuget` - (Optional) Deprecated since 6.9.0+ Nuget repository special configuration
  * `feed_context_path` - (Optional)
  * `download_context_path` - (Optional)
  * `v3_feed_url` - (Optional)
* `propagate_query_params` - (Optional, Generic repos only)
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
  * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
  * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
  * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
  * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'

## Import

Remote repositories can be imported using their name, e.g.

```
$ terraform import artifactory_remote_repository.my-remote my-remote
```
