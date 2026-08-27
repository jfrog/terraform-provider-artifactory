---
subcategory: "Remote Repositories"
---
# Artifactory Remote AI-Editor Extensions Repository Resource

Creates a remote AI-Editor Extensions repository that proxies and caches editor extensions from a VS Code compatible marketplace gallery, such as `https://marketplace.visualstudio.com/_apis/public/gallery`. This package type is modeled on VS Code marketplace extensions.

~> AI-Editor Extensions repositories are supported as **remote** repositories only, so the provider exposes no `artifactory_local_aieditorextensions_repository`, `artifactory_virtual_aieditorextensions_repository`, or `artifactory_federated_aieditorextensions_repository` resource.

-> `url` is required. Unlike some package types, Artifactory does not fall back to a default gallery URL for AI-Editor Extensions, so it must be set on every repository. Ex: Use `https://marketplace.visualstudio.com/_apis/public/gallery` for the VS Code marketplace.

## Example Usage

```hcl
resource "artifactory_remote_aieditorextensions_repository" "my-remote-aieditorextensions" {
  key         = "my-remote-aieditorextensions"
  url         = "https://marketplace.visualstudio.com/_apis/public/gallery"
  description = "AI-Editor (VS Code) extensions proxy"
}
```

Extension payloads are served from a CDN separate from the gallery host, so external dependency resolution is enabled by default for this package type. Override the patterns if your gallery serves payloads from a different host:

```hcl
resource "artifactory_remote_aieditorextensions_repository" "my-remote-aieditorextensions-custom" {
  key = "my-remote-aieditorextensions-custom"
  url = "https://marketplace.visualstudio.com/_apis/public/gallery"

  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/**vsassets.io/**", "**/**gallerycdn.vsassets.io/**"]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `url` - (Required) The URL of the marketplace gallery to proxy. Example: for the VS Code marketplace, use `https://marketplace.visualstudio.com/_apis/public/gallery`. Artifactory applies no default URL for this package type, so this attribute must always be set; omitting it fails with `No URL defined for remote repository`.
* `description` - (Optional)
* `notes` - (Optional)
* `external_dependencies_enabled` - (Optional) When set, Artifactory can resolve extension dependencies from the external sources matching `external_dependencies_patterns`. Unlike other remote repository types, this defaults to `true` for AI-Editor Extensions because extension payloads are hosted on a CDN separate from the gallery URL.
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote hosts external extension dependencies may be downloaded from. Only takes effect when `external_dependencies_enabled` is `true`, but Artifactory stores the patterns either way, so they may be set while it is `false`. Default value is `["**/**vsassets.io/**"]`. An empty list is not accepted — the provider requires at least one pattern.
* `bypass_head_requests` - (Read-only, always `true`) Artifactory always enables this setting for AI-Editor Extensions repositories, rather than `false` as on other remote repository types, and rejects any request that tries to change it. It is therefore exposed as a read-only attribute that always reports `true` and cannot be set in configuration (setting it — even to `true` — produces a "read-only attribute" error).
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication. Default value is `false`, matching the common remote default. OCI, Helm OCI, and Hugging Face remotes default to `true`.
* `propagate_query_params` - (Optional) When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository. Default value is `false`.
* `retrieve_sha256_from_server` - (Optional) When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo. Default value is `false`.
* `curated` - (Optional) Enable repository to be protected by the Curation service. Default value is `false`.
* `pass_through` - (Optional) Enable Pass-through for Curation Audit. When enabled, allows artifacts to pass through the Curation audit process. Default value is `false`.
* `custom_http_headers` - (Optional) Up to 5 custom HTTP headers sent on every outbound request to the remote URL. Requires Artifactory 7.146.0 or later. Header values are write-only: they are masked in plan output and never read back from Artifactory, so `terraform import` cannot recover them. To remove all headers, remove the attribute. Each entry supports:
    * `name` - (Required) Header name. Artifactory stores header names lower-cased.
    * `value` - (Required, Sensitive) Header value.
    * `sensitive` - (Optional) When `true`, Artifactory encrypts the value server-side. Default value is `false`.

```hcl
resource "artifactory_remote_aieditorextensions_repository" "my-remote-aieditorextensions-curated" {
  key = "my-remote-aieditorextensions-curated"
  url = "https://marketplace.visualstudio.com/_apis/public/gallery"

  curated      = true
  pass_through = false

  custom_http_headers = [
    { name = "x-api-key", value = "my-gallery-token", sensitive = true },
  ]
}
```

The default `repo_layout_ref` for this package type is `simple-default`, and `list_remote_folder_items` defaults to `false`.

~> Setting `enabled = true` inside the shared `content_synchronisation` block has no effect: Artifactory stores it as `false` regardless of what is sent, which leaves a perpetual diff in the plan. The nested `statistics_enabled`, `properties_enabled`, and `source_origin_absence_detection` flags do persist. This applies to non-smart remote repositories (those not proxying another Artifactory instance), which includes this package type.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_aieditorextensions_repository.my-remote-aieditorextensions my-remote-aieditorextensions
```
