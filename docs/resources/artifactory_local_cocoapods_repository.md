# Artifactory Local Cocoapods Repository Resource

Creates a local cocoapods repository.

## Example Usage

```hcl
resource "artifactory_local_cocoapods_repository" "terraform-local-test-cocoapods-repo" {
  key = "terraform-local-test-cocoapods-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Cocoapods repository type closely match with arguments for Generic repository type.
