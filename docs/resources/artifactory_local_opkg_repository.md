# Artifactory Local Opkg Repository Resource

Creates a local opkg repository.

## Example Usage

```hcl
resource "artifactory_local_opkg_repository" "terraform-local-test-opkg-repo" {
  key = "terraform-local-test-opkg-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Opkg repository type closely matches with arguments for Generic repository type.
