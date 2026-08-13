---
subcategory: "Remote Repositories"
---

# Artifactory Remote AI-Editor Extensions Repository Data Source

Retrieves configuration for a remote AI-Editor Extensions repository.

## Example Usage

```hcl
data "artifactory_remote_aieditorextensions_repository" "example" {
  key = "my-remote-aieditorextensions"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for remote repository data sources](remote.md). `package_type` is `aieditorextensions` and the default layout is `simple-default`.

In addition, the following type-specific attributes are exported:

* `external_dependencies_enabled` - When set, Artifactory can resolve extension dependencies from the external sources matching `external_dependencies_patterns`.
* `external_dependencies_patterns` - An allow list of Ant-style path patterns that determine which remote hosts external extension dependencies may be downloaded from.
* `enable_token_authentication` - Whether token (Bearer) based authentication is enabled.
* `propagate_query_params` - Whether query params included in the request to Artifactory are passed on to the remote repository.
* `retrieve_sha256_from_server` - Whether Artifactory retrieves the SHA256 from the remote server when it is not cached in the remote repo.
* `curated` - Whether the repository is protected by the Curation service.
* `pass_through` - Whether Pass-through for Curation Audit is enabled.

-> `custom_http_headers` is not exported. Artifactory returns header values in plaintext, and a data source has no configuration to compare them against, so exposing them would write secrets into state.
