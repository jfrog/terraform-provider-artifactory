---
subcategory: "Remote Repositories"
---
# Artifactory Remote JetBrains Plugins Repository Resource

Creates a remote JetBrains Plugins repository, which proxies and caches IDE plugins from the
JetBrains Marketplace (`https://plugins.jetbrains.com`).

~> This package type requires a **Pro** license and the *Packages IDE Extension* entitlement, and it must
be enabled on the instance. If either is missing, Artifactory rejects the create with
`The package type jetbrainsplugins is not supported`.

~> JetBrains Plugins is supported as a **remote** repository only, so the provider exposes no local,
virtual, or federated equivalent.

## Example Usage

```hcl
resource "artifactory_remote_jetbrainsplugins_repository" "my-remote-jetbrainsplugins" {
  key = "my-remote-jetbrainsplugins"
  url = "https://plugins.jetbrains.com"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `url` - (Required) The remote repo URL. Use `https://plugins.jetbrains.com` for the JetBrains Marketplace.
  Artifactory applies no default URL over the REST API and rejects the request with
  `No URL defined for remote repository` if it is omitted.
* `description` - (Optional)
* `notes` - (Optional)

`repo_layout_ref` defaults to `simple-default`.

There are no JetBrains Plugins specific attributes: the package type contributes only `url` to the
repository contract, so everything else comes from the common remote repository arguments.

All base remote attributes that don't appear in the excluded list below round-trip freely for
this type — including `bypass_head_requests`, `list_remote_folder_items`, `allow_any_host_auth`,
`enable_cookie_management`, `retrieval_cache_period_seconds`, `missed_cache_period_seconds`, and
the standard `blacked_out` / `offline` / `hard_fail` / `priority_resolution` / `xray_index` /
`property_sets` / `content_synchronisation` set. Unlike the AI-Editor Extensions repository,
`bypass_head_requests` is **not** pinned read-only for this type.

The following attributes are **not** part of the JetBrains Plugins contract and are not exposed —
Artifactory ignores or rejects them for this package type:

* `external_dependencies_enabled` / `external_dependencies_patterns` (not annotated for this type)
* `enable_token_authentication` (annotated for Docker / OCI / Helm OCI only)
* `curated` / `pass_through` (Curation is not supported for JetBrains Plugins; setting `curated = true` returns HTTP 400)

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_jetbrainsplugins_repository.my-remote-jetbrainsplugins my-remote-jetbrainsplugins
```
