---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Resource

Creates a remote Terraform repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

## Example Usage

```hcl
resource "artifactory_remote_terraform_repository" "terraform-remote" {
  key                     = "terraform-remote"
  url                     = "https://github.com/"
  terraform_registry_url  = "https://registry.terraform.io"
  terraform_providers_url = "https://releases.hashicorp.com"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The base URL of the Module storage API.
* `terraform_registry_url` - (Optional) The base URL of the registry API. 
  When using Smart Remote Repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey`.
* `terraform_providers_url` - (Optional) The base URL of the Provider's storage API.
  When using Smart remote repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey/providers`.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_terraform_repository.terraform-remote terraform-remote
```
