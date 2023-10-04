---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Conan Repository Data Source

Retrieves a virtual Conan repository.

## Example Usage

```hcl
data "artifactory_virtual_conan_repository" "virtual-conan" {
  key = "virtual-conan"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
* `force_conan_authentication` - Force basic authentication credentials in order to use this repository.
  Default is `false`.