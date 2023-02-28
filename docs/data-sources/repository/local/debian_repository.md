---
subcategory: "Local Repositories"
---

# Artifactory Local Debian Repository Data Source

Retrieves a local Debian repository.

## Example Usage

```hcl
data "artifactory_local_debian_repository" "local-test-debian-repo-basic" {
  key = "local-test-debian-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `primary_keypair_ref` - The primary RSA key to be used to sign packages.
* `secondary_keypair_ref` - The secondary RSA key to be used to sign packages.
* `index_compression_formats` - The options are Bzip2 (.bz2 extension) (default), LZMA (.lzma extension)
  and XZ (.xz extension).
* `trivial_layout` - When set, the repository will use the deprecated trivial layout.
