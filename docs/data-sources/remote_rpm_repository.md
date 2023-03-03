---
subcategory: "Remote Repositories"
---
# Artifactory Remote Rpm Repository Data Resource

Retrieves a remote Rpm repository.

## Example Usage

```hcl
data "artifactory_remote_rpm_repository" "my-remote-rpm" {
  key = "my-remote-rpm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):