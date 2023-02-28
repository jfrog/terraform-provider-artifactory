---
subcategory: "Local Repositories"
---

# Artifactory Local Nuget Repository Data Source

Retrieves a local Nuget repository.

## Example Usage

```hcl
data "artifactory_local_nuget_repository" "local-test-nuget-repo-basic" {
  key = "local-test-nuget-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `max_unique_snapshots` - The maximum number of unique snapshots of a single artifact to store Once the
  number of snapshots exceeds this setting, older versions are removed A value of 0 (default) indicates there is no
  limit, and unique snapshots are not cleaned up.
* `force_nuget_authentication` - Force basic authentication credentials in order to use this repository.
  Default is `false`.
