---
subcategory: "Virtual Repositories"
---

# Artifactory Virtual Nix Repository Data Source

Retrieves configuration for a virtual Nix repository.

## Example Usage

```hcl
data "artifactory_virtual_nix_repository" "example" {
  key = "my-nix-virtual"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for virtual repository data sources](virtual.md). `package_type` is `nix` and the default layout is `nix-default`.
