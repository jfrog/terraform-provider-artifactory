---
subcategory: "Local Repositories"
---

# Artifactory Local Alpine Repository Data Source

Retrieves a local alpine repository.

## Example Usage

```hcl
data "artifactory_local_alpine_repository" "local-test-alpine-repo-basic" {
  key = "local-test-alpine-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `primary_keypair_ref` - The RSA key to be used to sign alpine indices.
