---
subcategory: "Remote Repositories"
---

# Artifactory Remote Nix Repository Data Source

Retrieves configuration for a remote Nix repository.

## Example Usage

```hcl
data "artifactory_remote_nix_repository" "example" {
  key = "my-nix-remote"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for remote repository data sources](remote.md). `package_type` is `nix` and the default layout is `nix-default`.
