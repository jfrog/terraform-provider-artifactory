---
subcategory: "Remote Repositories"
---
# Artifactory Remote NVIDIA NIM Repository Resource

Creates a remote NVIDIA NIM repository that proxies and caches NIM models from an NGC-compatible model registry, such as `https://api.ngc.nvidia.com`.

~> NVIDIA NIM repositories are supported as **remote** repositories only, so the provider exposes no `artifactory_local_nimmodel_repository`, `artifactory_virtual_nimmodel_repository`, or `artifactory_federated_nimmodel_repository` resource.

-> `url` is required. Artifactory does not fall back to a default registry URL for NVIDIA NIM, so it must be set on every repository. Ex: use `https://api.ngc.nvidia.com` for the public NVIDIA NGC catalog.

## Example Usage

```hcl
resource "artifactory_remote_nimmodel_repository" "my-remote-nimmodel" {
  key         = "my-remote-nimmodel"
  url         = "https://api.ngc.nvidia.com"
  description = "NVIDIA NIM models proxy"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `url` - (Required) The URL of the NVIDIA NIM registry to proxy. Example: for the public NVIDIA NGC catalog, use `https://api.ngc.nvidia.com`. Artifactory applies no default URL for this package type, so this attribute must always be set; omitting it fails with `No URL defined for remote repository`.
* `description` - (Optional)
* `notes` - (Optional)
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication. Default value is `true`, matching the Artifactory default for this package type. Note this differs from most other remote repository types, which default it to `false`; the Docker, OCI, and HelmOCI remote repository resources share this same `true` default.

The default `repo_layout_ref` for this package type is `maven-2-default`, and `list_remote_folder_items` defaults to `false`.

~> Setting `enabled = true` inside the shared `content_synchronisation` block has no effect: Artifactory stores it as `false` regardless of what is sent, which leaves a perpetual diff in the plan. The nested `statistics_enabled`, `properties_enabled`, and `source_origin_absence_detection` flags do persist. This applies to non-smart remote repositories (those not proxying another Artifactory instance), which includes this package type.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_nimmodel_repository.my-remote-nimmodel my-remote-nimmodel
```
