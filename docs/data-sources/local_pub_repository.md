---
subcategory: "Local Repositories"
---

# Artifactory Local Pub Repository Data Source

Retrieves a local pub repository.

## Example Usage

```hcl
data "artifactory_local_pub_repository" "local-test-pub-repo" {
  key = "local-test-pub-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
