---
subcategory: "Remote Repositories"
---
# Artifactory Remote Terraform Repository Data Source

Retrieves a remote Terraform repository.

## Example Usage

```hcl
data "artifactory_remote_terraform_repository" "remote-terraform" {
  key = "remote-terraform"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `terraform_registry_url` - (Optional) The base URL of the registry API. When using Smart Remote Repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey`.
* `terraform_providers_url` - (Optional) The base URL of the Provider's storage API. When using Smart remote repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey/providers`.
