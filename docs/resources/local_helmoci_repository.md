---
subcategory: "Local Repositories"
---
# Artifactory Local Helm OCI Repository Resource

Creates a local Helm OCI repository.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/helm-oci-repositories)

## Example Usage

```hcl
resource "artifactory_local_helmoci_repository" "my-helmoci-local" {
  key 	          = "my-helmoci-local"
  tag_retention   = 3
  max_unique_tags = 5
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `tag_retention` - (Optional) If greater than 1, overwritten tags will be saved by their digest, up to the set up number.
* `max_unique_tags` - (Optional) The maximum number of unique tags of a single OCI image to store in this 
repository. Once the number tags for an image exceeds this setting, older tags are removed. 
A value of 0 (default) indicates there is no limit.

## Import

Local repositories can be imported using their name, e.g.

```
$ terraform import artifactory_local_helmoci_repository.my-helmoci-local my-helmoci-local
```
