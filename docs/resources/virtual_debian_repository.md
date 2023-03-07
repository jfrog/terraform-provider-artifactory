---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Debian Repository Resource

Creates a virtual Debian repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Debian+Repositories#DebianRepositories-VirtualRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_debian_repository" "foo-debian" {
  key                                 = "foo-debian"
  repositories                        = []
  description                         = "A test virtual repo"
  notes                               = "Internal description"
  includes_pattern                    = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                    = "com/google/**"
  optional_index_compression_formats  = [ "bz2", "xz" ]
  debian_default_architectures        = "amd64,i386"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
* `primary_keypair_ref` - (Optional) Primary keypair used to sign artifacts. Default is empty.
* `secondary_keypair_ref` - (Optional) Secondary keypair used to sign artifacts. Default is empty.
* `optional_index_compression_formats` - (Optional) Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are `bz2`,`lzma` and `xz`. Default value is `bz2`.
* `debian_default_architectures` - (Optional) Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_debian_repository.foo-debian foo-debian
```
