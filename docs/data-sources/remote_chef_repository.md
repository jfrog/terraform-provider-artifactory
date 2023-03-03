---
subcategory: "Remote Repositories"
---
# Artifactory Remote Chef Repository Data Resource

Retrieves a remote Chef repository.

## Example Usage

```hcl
data "artifactory_remote_chef_repository" "my-remote-chef" {
  key = "my-remote-chef"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):