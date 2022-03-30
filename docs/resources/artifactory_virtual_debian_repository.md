# Artifactory Virtual Debian Repository Resource

Provides an Artifactory virtual repository resource with specific debian features.

## Example Usage

```hcl
resource "artifactory_virtual_debian_repository" "foo-debian" {
  key          = "foo-debian"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  optional_index_compression_formats = [ "bz2", "xz" ]
  debian_default_architectures = "amd64,i386"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `repositories` - (Required, but may be empty)
* `description` - (Optional)
* `notes` - (Optional)
* `primary_keypair_ref` - (Optional) Primary keypair used to sign artifacts. Default is empty.
* `secondary_keypair_ref` - (Optional) Secondary keypair used to sign artifacts. Default is empty.
* `optional_index_compression_formats` - (Optional) Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are 'bz2','lzma' and 'xz'. Default value is 'bz2'.
* `debian_default_architectures` - (Optional) Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.

Arguments for Debian repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_debian_repository.foo foo
```
