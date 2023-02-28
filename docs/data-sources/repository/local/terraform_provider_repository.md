---
subcategory: "Local Repositories"
---

# Artifactory Local Terraform Provider Repository Data Source

Retrieves a local Terraform Provider repository. Official documentation can be
found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

## Example Usage

```hcl
data "artifactory_local_terraform_provider_repository" "local-test-terraform-provider-repo" {
  key = "local-test-terraform-provider-repo"
}
```

## Argument Reference

* `key` - (Required) the identity key of the repo.

## Attribute Reference

Attributes have a one to one mapping with
the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following attributes are
supported, along with the [common list of arguments for the local repositories](local.md):

* `description` - (Optional)
* `notes` - (Optional)
