---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Bower Repository Resource

Creates a virtual Bower repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Bower+Repositories#BowerRepositories-VirtualRepositories).
## Example Usage

```hcl
resource "artifactory_virtual_bower_repository" "foo-bower" {
  key                           = "foo-bower"
  repositories                  = []
  description                   = "A test virtual repo"
  notes                         = "Internal description"
  includes_pattern              = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern              = "com/google/**"
  external_dependencies_enabled = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten. Default value is false.
* `external_dependencies_remote_repo` - (Optional) The remote repository aggregated by this virtual repository in which the external dependency will be cached.
* `external_dependencies_patterns` - (Optional) An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. By default, this is set to ** which means that dependencies may be downloaded from any external source.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_bower_repository.foo-bower foo-bower
```
