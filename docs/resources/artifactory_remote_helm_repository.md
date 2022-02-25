# Artifactory Remote Repository Resource

Provides an Artifactory remote `helm` repository resource. This provides helm specific fields and is the only way to get them.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Package+Management),
although helm is (currently) not listed as a supported format

## Example Usage
Includes only new and relevant fields, for anything else, see: [generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_helm_repository" "helm-remote" {
  key = "helm-remote-foo25"
  url = "https://repo.chartcenter.io/"
  helm_charts_base_url = "https://foo.com"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**github.com**"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
All generic repo arguments are supported, in addition to:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `project_key` - (Optional) Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: "DEV" or "PROD"
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `username` - (Optional)
* `password` - (Optional)
* `proxy` - (Optional)
* `includes_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).
* `excludes_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.
* `repo_layout_ref` - (Optional) Repository layout key for the remote repository
* `remote_repo_layout_ref` - (Optional) Repository layout key for the remote layout mapping
* `hard_fail` - (Optional)
* `offline` - (Optional) If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.
* `blacked_out` - (Optional) (A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.
* `helm_charts_base_url` - (Optional) - No documentation is available. Hopefully you know what this means
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten.
* `external_dependencies_patterns` - (Optional) An Allow List of Ant-style path expressions that specify where external
  dependencies may be downloaded from. By default, this is an empty list which means that no dependencies may be downloaded
  from external sources. Note that the official documentation states the default is '**', which is correct when creating
  repositories in the UI, but incorrect for the API.
* `xray_index` - (Optional, Default: false)  Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.
* `store_artifacts_locally` - (Optional) When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming it directly to satellite pass-though Artifactory servers.
* `socket_timeout_millis` - (Optional) Network timeout (in ms) to use when establishing a connection and for unanswered requests. Timing out on a network operation is considered a retrieval failure.
* `local_address` - (Optional) The local address to be used when creating connections. Useful for specifying the interface to use on systems with multiple network interfaces.
* `retrieval_cache_period_seconds` - (Optional, Default: 7200) The metadataRetrievalTimeoutSecs field not allowed to be bigger then retrievalCachePeriodSecs field.
* `failed_retrieval_cache_period_secs` - (Optional) This field is not returned in a get payload but is offered on the UI. It's inserted here for inclusive and informational reasons. It does not function
* `missed_cache_period_seconds` - (Optional) The number of seconds to cache artifact retrieval misses (artifact not found). A value of 0 indicates no caching.
* `unused_artifacts_cleanup_period_enabled` - (Optional)
* `unused_artifacts_cleanup_period_hours` - (Optional) The number of hours to wait before an artifact is deemed "unused" and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.
* `assumed_offline_period_secs` - (Optional, Default: 300) The number of seconds the repository stays in assumed offline state after a connection error. At the end of this time, an online check is attempted in order to reset the offline status. A value of 0 means the repository is never assumed offline. Default to 300.
* `share_configuration` - (Optional)
* `synchronize_properties` - (Optional) When set, remote artifacts are fetched along with their properties.
* `block_mismatching_mime_types` - (Optional) Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.
* `property_sets` - (Optional) List of property set name
* `allow_any_host_auth` - (Optional) Also known as 'Lenient Host Authentication', Allow credentials of this repository to be used on requests redirected to any other host.
* `enable_cookie_management` - (Optional) Enables cookie management if the remote repository uses cookies to manage client state.
* `bypass_head_requests` - (Optional) Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.
* `priority_resolution` - (Optional) Setting repositories with priority will cause metadata to be merged only from repositories set with this field
* `client_tls_certificate` - (Optional)
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
  * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
  * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
  * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
  * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'
* `xray_index` - (Optional, Default: false)  Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.
