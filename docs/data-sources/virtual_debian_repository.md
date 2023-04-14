---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Debian Repository Data Source

Retrieves a virtual Debian repository.

## Example Usage

```hcl
data "artifactory_virtual_debian_repository" "virtual-debian" {
  key = "virtual-debian"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
* `primary_keypair_ref` - (Optional) Primary keypair used to sign artifacts. Default is empty.
* `secondary_keypair_ref` - (Optional) Secondary keypair used to sign artifacts. Default is empty.
* `optional_index_compression_formats` - (Optional) Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are `bz2`,`lzma` and `xz`. Default value is `bz2`.
* `debian_default_architectures` - (Optional) Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.
