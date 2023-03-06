---
subcategory: "Remote Repositories"
---
# Artifactory Remote Terraform Repository Data Source

Retrieves a remote Terraform repository.

## Example Usage

```hcl
data "artifactory_remote_terraform_repository" "my-remote-terraform" {
  key = "my-remote-terraform"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):