---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Alpine Repository Data Source

Retrieves a virtual Alpine repository.

## Example Usage

```hcl
data "artifactory_virtual_alpine_repository" "virtual-alpine" {
  key = "virtual-alpine"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/virtual.md):

* `primary_keypair_ref` - (Optional) Primary keypair used to sign artifacts. Default value is empty.
* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
