---
subcategory: "Local Repositories"
---
# Artifactory Local Terraform Module Repository Resource

Creates a local Terraform Module repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

## Example Usage

```hcl
resource "artifactory_local_terraform_module_repository" "terraform-local-test-terraform-module-repo" {
  key = "terraform-local-test-terraform-module-repo"
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
$ terraform import artifactory_local_terraform_module_repository.terraform-local-test-terraform-module-repo terraform-local-test-terraform-module-repo
```
