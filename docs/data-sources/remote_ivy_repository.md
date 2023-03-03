---
subcategory: "Remote Repositories"
---
# Artifactory Remote Ivy Repository Data Resource

Retrieves a remote Ivy repository.

## Example Usage

```hcl
data "artifactory_remote_ivy_repository" "my-remote-ivy" {
  key = "my-remote-ivy"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):