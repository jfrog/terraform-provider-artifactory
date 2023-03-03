---
subcategory: "Remote Repositories"
---
# Artifactory Remote NuGet Repository Data Resource

Retrieves a remote NuGet repository.

## Example Usage

```hcl
data "artifactory_remote_nuget_repository" "my-remote-nuget" {
  key = "my-remote-nuget"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):