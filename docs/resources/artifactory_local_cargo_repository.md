# Artifactory Local Cargo Repository Resource

Creates a local Cargo repository

## Example Usage

```hcl
resource "artifactory_local_cargo_repository" "terraform-local-test-cargo-repo-basic" {
  key                        = "terraform-local-test-cargo-repo-basic"
  anonymous_access           = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `anonymous_access` - (Optional) Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is 'false'.

Arguments for Cargo repository type closely match with arguments for Generic repository type.
