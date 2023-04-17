---
subcategory: "Remote Repositories"
---
# Artifactory Remote Rpm Repository Data Source

Retrieves a remote Rpm repository.

## Example Usage

```hcl
data "artifactory_remote_rpm_repository" "remote-rpm" {
  key = "remote-rpm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.