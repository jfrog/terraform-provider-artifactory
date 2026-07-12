---
subcategory: "Local Repositories"
---

# Artifactory Local Nix Repository Data Source

Retrieves configuration for a local Nix repository.

## Example Usage

```hcl
data "artifactory_local_nix_repository" "example" {
  key = "my-nix-local"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for local repository data sources](local.md). `package_type` is `nix` and the default layout is `nix-default`.
