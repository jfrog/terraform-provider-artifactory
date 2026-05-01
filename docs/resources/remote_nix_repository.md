---
subcategory: "Remote Repositories"
---
# Artifactory Remote Nix Repository Resource

Creates a remote Nix repository that proxies and caches a Nix binary cache (for example `https://cache.nixos.org`). See [Nix repositories](https://docs.jfrog.com/artifactory/docs/nix-repositories).

## Example Usage

```hcl
resource "artifactory_remote_nix_repository" "my-nix-remote" {
  key         = "my-nix-remote"
  url         = "https://cache.nixos.org"
  description = "Remote Nix cache"
}
```

## Argument Reference

The following arguments are supported, along with the [common arguments for remote repositories](remote.md):

* `key` - (Required) Repository key.
* `url` - (Required) URL of the upstream Nix binary cache.
* `repo_layout_ref` - (Optional) Repository layout reference. Defaults to `nix-default` when unset.

## Import

Repositories can be imported using the repository key, for example:

```
$ terraform import artifactory_remote_nix_repository.my-nix-remote my-nix-remote
```
