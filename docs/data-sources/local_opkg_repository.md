---
subcategory: "Local Repositories"
---

# Artifactory Local OPKG Repository Data Source

Retrieves a local opkg repository.

## Example Usage

```hcl
data "artifactory_local_opkg_repository" "local-test-opkg-repo" {
  key = "local-test-opkg-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
