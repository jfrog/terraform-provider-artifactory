---
subcategory: "Remote Repositories"
---
# Artifactory Remote Ansible Repository Data Source

Retrieves a remote Ansible repository.

## Example Usage

```hcl
data "artifactory_remote_ansible_repository" "remote-ansible" {
  key = "remote-ansible"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.