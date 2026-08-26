---
subcategory: "Remote Repositories"
---
# Artifactory Remote JetBrains Plugins Repository Data Source

Retrieves a remote JetBrains Plugins repository.

## Example Usage

```hcl
data "artifactory_remote_jetbrainsplugins_repository" "remote-jetbrainsplugins" {
  key = "remote-jetbrainsplugins"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.

The write-only credential attributes (`password`, `password_wo`) are deliberately not exported, so no
secret is written to state by this data source.
