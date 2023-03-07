---
subcategory: "Remote Repositories"
---
# Artifactory Remote Alpine Repository Data Source

Retrieves a remote Alpine repository.

## Example Usage

```hcl
data "artifactory_remote_alpine_repository" "my-remote-alpine" {
  key = "my-remote-alpine"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](remote.md) are supported.