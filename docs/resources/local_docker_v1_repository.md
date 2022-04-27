---
subcategory: "Local Repositories"
---
# Artifactory Local Docker V1 Repository Resource

Creates a local Docker v1 repository - By choosing a V1 repository, you don't really have many options.

## Example Usage

```hcl
resource "artifactory_local_docker_v1_repository" "foo" {
  key = "foo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_docker_v1_repository.foo foo
```
