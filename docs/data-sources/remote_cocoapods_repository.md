---
subcategory: "Remote Repositories"
---
# Artifactory Remote CocoaPods Repository Data Resource

Retrieves a remote CocoaPods repository.

## Example Usage

```hcl
data "artifactory_remote_cocoapods_repository" "my-remote-cocoapods" {
  key = "my-remote-cocoapods"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):