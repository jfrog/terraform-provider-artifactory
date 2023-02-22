---
subcategory: "Local Repositories"
---

# Artifactory Local Alpine Repository Data Source

Provides a data source for alpine repositories.

## Example Usage

```hcl
data "artifactory_local_alpine_repository" "terraform-local-test-alpine-repo-basic" {
  key                 = "terraform-local-test-alpine-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `primary_keypair_ref` - The RSA key to be used to sign alpine indices.
