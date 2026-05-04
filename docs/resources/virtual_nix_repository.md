---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Nix Repository Resource

Creates a virtual Nix repository that aggregates local and remote Nix repositories behind a single URL. See [Nix repositories](https://docs.jfrog.com/artifactory/docs/nix-repositories).

## Example Usage

```hcl
resource "artifactory_local_nix_repository" "local" {
  key = "nix-local"
}

resource "artifactory_remote_nix_repository" "remote" {
  key = "nix-remote"
  url = "https://cache.nixos.org"
}

resource "artifactory_virtual_nix_repository" "nix" {
  key            = "nix-virtual"
  repositories   = [
    artifactory_local_nix_repository.local.key,
    artifactory_remote_nix_repository.remote.key,
  ]
}
```

## Argument Reference

The following arguments are supported, along with the [common arguments for virtual repositories](virtual.md):

* `key` - (Required) Repository key.
* `repositories` - (Optional) Ordered list of repository keys included in the virtual repository.
* `repo_layout_ref` - (Optional) Repository layout reference. Defaults to `nix-default` when unset.

## Import

Repositories can be imported using the repository key, for example:

```
$ terraform import artifactory_virtual_nix_repository.nix nix-virtual
```
