---
subcategory: "Remote Repositories"
---
# Artifactory Remote Maven Repository Data Resource

Retrieves a remote Maven repository.

## Example Usage

```hcl
data "artifactory_remote_maven_repository" "my-remote-maven" {
  key = "my-remote-maven"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):