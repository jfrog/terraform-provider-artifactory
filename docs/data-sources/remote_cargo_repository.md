---
subcategory: "Remote Repositories"
---
# Artifactory Remote Cargo Repository Data Resource

Retrieves a remote Cargo repository.

## Example Usage

```hcl
data "artifactory_remote_cargo_repository" "my-remote-cargo" {
  key = "my-remote-cargo"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):