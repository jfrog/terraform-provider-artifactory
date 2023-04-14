---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Maven Repository Data Source

Retrieves a virtual Maven repository.

## Example Usage

```hcl
data "artifactory_virtual_maven_repository" "virtual-maven" {
  key = "virtual-maven"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `pom_repository_references_cleanup_policy` - (Optional) One of: `"discard_active_reference", "discard_any_reference", "nothing"`
* `force_maven_authentication` - (Optional) Forces authentication when fetching from remote repos.
