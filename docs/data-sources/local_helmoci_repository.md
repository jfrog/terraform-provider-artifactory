---
subcategory: "Local Repositories"
---

# Artifactory Local Helm OCI Repository Data Source

Retrieves a local Helm OCI repository resource

## Example Usage

```hcl
resource "artifactory_local_helmoci_repository" "my-helmoci-local" {
  key = "my-helmoci-local"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `tag_retention` - If greater than 1, overwritten tags will be saved by their digest, up to the set up number.
* `max_unique_tags` - The maximum number of unique tags of a single Docker image to store in this repository. Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.
