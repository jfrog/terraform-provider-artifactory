---
subcategory: "Remote Repositories"
---
# Artifactory Remote Docker Repository Data Source

Retrieves a remote Docker repository.

## Example Usage

```hcl
data "artifactory_remote_docker_repository" "my-remote-docker" {
  key = "my-remote-docker"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):