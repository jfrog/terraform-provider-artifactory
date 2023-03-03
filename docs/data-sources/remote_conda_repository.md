---
subcategory: "Remote Repositories"
---
# Artifactory Remote Conda Repository Data Resource

Retrieves a remote Conda repository.

## Example Usage

```hcl
data "artifactory_remote_conda_repository" "my-remote-conda" {
  key = "my-remote-conda"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):