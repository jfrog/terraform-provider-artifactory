---
subcategory: "Remote Repositories"
---
# Artifactory Remote Helm Repository Data Resource

Retrieves a remote Helm repository.

## Example Usage

```hcl
data "artifactory_remote_helm_repository" "my-remote-helm" {
  key = "my-remote-helm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):