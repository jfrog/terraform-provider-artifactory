---
subcategory: "Remote Repositories"
---
# Artifactory Remote Swift Repository Data Source

Retrieves a remote Swift repository.

## Example Usage

```hcl
data "artifactory_remote_swift_repository" "my-remote-swift" {
  key = "my-remote-swift"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):