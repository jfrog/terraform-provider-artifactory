---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Docker Repository Resource

Creates a virtual Docker repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Docker+Registry#DockerRegistry-VirtualDockerRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_docker_repository" "foo-docker" {
  key                               = "foo-docker"
  repositories                      = []
  description                       = "A test virtual repo"
  notes                             = "Internal description"
  includes_pattern                  = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                  = "com/google/**"
  resolve_docker_tags_by_timestamp  = true
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
* `resolve_docker_tags_by_timestamp` - (Optional) When enabled, in cases where the same Docker tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp. Default values is `false`.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_docker_repository.foo-docker foo-docker
```
