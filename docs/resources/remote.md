---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Common Arguments

The list of arguments, common for the remote repositories. All these arguments can be used together with the
repository-specific arguments, listed in separate repository-specific documents.  

### Passwords
Passwords can only be used when encryption is turned off, according to [Key Encryption instruction](https://www.jfrog.com/confluence/display/RTF/Artifactory+Key+Encryption).
Since only the artifactory server can decrypt them it is impossible for terraform to diff changes correctly.

To get full management, passwords can be decrypted globally using `POST /api/system/decrypt`. If this is not possible,
the password diff can be disabled per resource with-- noting that this will require resources to be tainted for an update:

```hcl
lifecycle {
    ignore_changes = ["password"]
}
```

## Example Usage (generic repository type)

```hcl
resource "artifactory_remote_generic_repository" "my-remote-generic" {
  key                         = "my-remote-generic"
  url                         = "http://testartifactory.io/artifactory/example-generic/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported:

All generic repo arguments are supported, in addition to:
* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `description` - (Optional) Public description.
* `notes` - (Optional) Internal description.
* `project_key` - (Optional) Project key for assigning this repository to. Must be 2 - 20 lowercase alphanumeric and hyphen characters.
  When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
  We don't recommend using this attribute to assign the repository to the project. Use the `repos` attribute in Project provider
  to manage the list of repositories. Default value - `default`.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: `DEV` or `PROD`.
  Before Artifactory 7.53.1, up to 2 values (`DEV` and `PROD`) are allowed. From 7.53.1 onward, only one value is allowed.
  The attribute should only be used if the repository is already assigned to the existing project.
  If not, the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create state
  drift during the update.
* `url` - (Required) The remote repo URL.
* `username` - (Optional)
* `password` - (Optional)
* `proxy` - (Optional) Proxy key from Artifactory Proxies settings. Default is empty field.
* `includes_pattern` - (Optional, Default: `**/*`) List of comma-separated artifact patterns to include when evaluating artifact requests in the form of x/y/\**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included.
* `excludes_pattern` - (Optional) List of comma-separated artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.
* `repo_layout_ref` - (Optional) Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.
* `remote_repo_layout_ref` - (Optional) Deprecated field. This field has currently no effect, because there is no corresponding field in the API body, and it's not returned by the GET call.
* `hard_fail` - (Optional, Default: `false`) When set, Artifactory will return an error to the client that causes the build to fail if there is a failure to communicate with this repository.
* `offline` - (Optional, Default: `false`) If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.
* `blacked_out` - (Optional) (A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution. Default is `false`.
* `xray_index` - (Optional, Default: `false`) Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.  Default is `false`.
* `store_artifacts_locally` - (Optional, Default: `false`) When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming it directly to satellite pass-though Artifactory servers.
* `socket_timeout_millis` - (Optional, Default: `15000`) Network timeout (in ms) to use when establishing a connection and for unanswered requests. Timing out on a network operation is considered a retrieval failure.
* `local_address` - (Optional) The local address to be used when creating connections. Useful for specifying the interface to use on systems with multiple network interfaces.
* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) The metadataRetrievalTimeoutSecs field not allowed to be bigger then retrievalCachePeriodSecs field.
* `missed_cache_period_seconds` - (Optional, Default: `1800`) The number of seconds to cache artifact retrieval misses (artifact not found). A value of 0 indicates no caching.
* `unused_artifacts_cleanup_period_hours` - (Optional, Default: `0`) Unused Artifacts Cleanup Period (Hr) in the UI. The number of hours to wait before an artifact is deemed 'unused' and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.
* `assumed_offline_period_secs` - (Optional, Default: `300`) The number of seconds the repository stays in assumed offline state after a connection error. At the end of this time, an online check is attempted in order to reset the offline status. A value of 0 means the repository is never assumed offline.
* `share_configuration` - (Optional) The attribute is 'Computed', so it's not managed by the Provider. There is no corresponding field in the UI, but the attribute is returned by Get.
* `synchronize_properties` - (Optional, Default: `false`) When set, remote artifacts are fetched along with their properties.
* `block_mismatching_mime_types` - (Optional, Default: `true`) Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request. Note: dafault value in the UI is `true`, but it is `false` if the repo was created using the API call. We are copying the UI behavior.
* `mismatching_mime_types_override_list` - (Optional) The set of mime types that should override the block_mismatching_mime_types setting. Eg: "application/json,application/xml". Default value is empty.
* `property_sets` - (Optional) List of property set names.
* `allow_any_host_auth` - (Optional, Default: `false`) Also known as 'Lenient Host Authentication', Allow credentials of this repository to be used on requests redirected to any other host.
* `enable_cookie_management` - (Optional, Default: `false`) Enables cookie management if the remote repository uses cookies to manage client state.
* `bypass_head_requests` - (Optional, Default: `false`) Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.
* `priority_resolution` - (Optional, Default: `false`) Setting repositories with priority will cause metadata to be merged only from repositories set with this field.
* `client_tls_certificate` - (Optional) Client TLS certificate name.
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories).
  * `enabled` - (Optional, Default: `false`) If set, Remote repository proxies a local or remote repository from another instance of Artifactory.
  * `statistics_enabled` - (Optional, Default: `false`) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain.
  * `properties_enabled` - (Optional, Default: `false`) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance.
  * `source_origin_absence_detection` - (Optional, Default: `false`) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance.
* `query_params` - (Optional) Custom HTTP query parameters that will be automatically included in all remote resource requests. For example: "param1=val1&param2=val2&param3=val3"
* `list_remote_folder_items` - (Optional, Default: `false`) Lists the items of remote folders in simple and list browsing. The remote content is cached according to the value of the 'Retrieval Cache Period'. This field exists in the API but not in the UI.
* `download_direct` - (Optional, Default: `false`) When set, download requests to this repository will redirect the client to download 
the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.
* `cdn_redirect` - (Optional) When set, download requests to this repository will redirect the client to download
the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only.
