---
subcategory: "Local Repositories"
---

# Artifactory Local Chef Repository Data Source

Retrieves a local Chef repository.

## Example Usage

```hcl
data "artifactory_local_chef_repository" "local-test-chef-repo" {
  key = "local-test-chef-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)
