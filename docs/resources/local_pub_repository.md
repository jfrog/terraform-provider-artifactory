---
subcategory: "Local Repositories"
---
# Artifactory Local Pub Repository Resource

Creates a local Pub repository.

## Example Usage

```hcl
resource "artifactory_local_pub_repository" "terraform-local-test-pub-repo" {
  key = "terraform-local-test-pub-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_pub_repository.terraform-local-test-pub-repo terraform-local-test-pub-repo
```
