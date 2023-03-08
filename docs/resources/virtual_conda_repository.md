---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Conda Repository Resource

Creates a virtual Conda repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Conda+Repositories#CondaRepositories-VirtualRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_conda_repository" "foo-conda" {
  key               = "foo-conda"
  repositories      = []
  description       = "A test virtual repo"
  notes             = "Internal description"
  includes_pattern  = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern  = "com/google/**"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_conda_repository.foo-conda foo-conda
```
