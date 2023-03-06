---
subcategory: "Remote Repositories"
---
# Artifactory Remote Composer Repository Data Source

Retrieves a remote Composer repository.

## Example Usage

```hcl
data "artifactory_remote_composer_repository" "my-remote-composer" {
  key = "my-remote-composer"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):