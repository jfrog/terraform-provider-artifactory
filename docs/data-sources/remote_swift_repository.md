---
subcategory: "Remote Repositories"
---
# Artifactory Remote Swift Repository Data Source

Retrieves a remote Swift repository.

## Example Usage

```hcl
data "artifactory_remote_swift_repository" "remote-swift" {
  key = "remote-swift"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.