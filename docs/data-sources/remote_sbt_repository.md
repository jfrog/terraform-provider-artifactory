---
subcategory: "Remote Repositories"
---
# Artifactory Remote SBT Repository Data Source

Retrieves a remote SBT repository.

## Example Usage

```hcl
data "artifactory_remote_sbt_repository" "my-remote-sbt" {
  key = "my-remote-sbt"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):