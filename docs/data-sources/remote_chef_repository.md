---
subcategory: "Remote Repositories"
---
# Artifactory Remote Chef Repository Data Source

Retrieves a remote Chef repository.

## Example Usage

```hcl
data "artifactory_remote_chef_repository" "remote-chef" {
  key = "remote-chef"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.