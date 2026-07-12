---
subcategory: "Local Repositories"
---
# Artifactory Local Nix Repository Resource

Creates a local Nix repository for hosting Nix store artifacts and channels. See [Nix repositories](https://docs.jfrog.com/artifactory/docs/nix-repositories) in the JFrog documentation.

## Example Usage

```hcl
resource "artifactory_local_nix_repository" "my-nix-local" {
  key         = "my-nix-local"
  description = "Local Nix repository"
  notes       = "Internal repository"
}
```

## Argument Reference

The following arguments are supported, along with the [common arguments for local repositories](local.md):

* `key` - (Required) Repository key.
* `repo_layout_ref` - (Optional) Repository layout reference. Defaults to `nix-default` when unset.

## Import

Repositories can be imported using the repository key, for example:

```
$ terraform import artifactory_local_nix_repository.my-nix-local my-nix-local
```
