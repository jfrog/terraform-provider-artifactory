---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Ivy Repository Data Source

Retrieves a virtual Ivy repository.

## Example Usage

```hcl
data "artifactory_virtual_ivy_repository" "virtual-ivy" {
  key = "virtual-ivy"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `pom_repository_references_cleanup_policy` - (Optional)
    - (1: discard_active_reference) Discard Active References - Removes repository elements that are declared directly under project or under a profile in the same POM that is activeByDefault.
    - (2: discard_any_reference) Discard Any References - Removes all repository elements regardless of whether they are included in an active profile or not.
    - (3: nothing) Nothing - Does not remove any repository elements declared in the POM.
* `key_pair` - (Optional) The keypair used to sign artifacts.
