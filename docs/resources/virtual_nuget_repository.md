---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Nuget Repository Resource

Creates a virtual Nuget repository. 
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/NuGet+Repositories#NuGetRepositories-VirtualRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_nuget_repository" "foo-nuget" {
  key                         = "foo-nuget"
  repositories                = []
  description                 = "A test virtual repo"
  notes                       = "Internal description"
  includes_pattern            = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern            = "com/google/**"
  force_nuget_authentication  = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `force_nuget_authentication` - (Optional) If set, user authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests. Default is `false`.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_nuget_repository.foo-nuget foo-nuget
```
