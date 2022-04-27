---
subcategory: "Local Repositories"
---
# Artifactory Local Docker V2 Repository Resource

Creates a local Docker v2 repository.

## Example Usage

```hcl
resource "artifactory_local_docker_v2_repository" "foo" {
  key 	          = "foo"
  tag_retention   = 3
  max_unique_tags = 5
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `block_pushing_schema1` - (Optional) When set, Artifactory will block the pushing of Docker images with manifest 
v2 schema 1 to this repository.
* `tag_retention` - (Optional) If greater than 1, overwritten tags will be saved by their digest, up to the set up 
number. This only applies to manifest V2.
* `max_unique_tags` - (Optional) The maximum number of unique tags of a single Docker image to store in this 
repository. Once the number tags for an image exceeds this setting, older tags are removed. 
A value of 0 (default) indicates there is no limit. This only applies to manifest v2.



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_docker_v2_repository.foo foo
```
