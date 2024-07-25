---
subcategory: "Local Repositories"
---

# Artifactory Local Ansible Repository Data Source

Retrieves a local Ansible repository.

## Example Usage

```hcl
data "artifactory_local_ansible_repository" "local-test-ansible-repo-basic" {
  key = "local-test-ansible-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `primary_keypair_ref` - The RSA key to be used to sign Ansible indices.
