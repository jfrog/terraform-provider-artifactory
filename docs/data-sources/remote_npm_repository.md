---
subcategory: "Remote Repositories"
---
# Artifactory Remote Npm Repository Data Resource

Retrieves a remote Npm repository.

## Example Usage

```hcl
data "artifactory_remote_npm_repository" "my-remote-npm" {
  key = "my-remote-npm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):