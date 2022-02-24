# Artifactory Local Gems Repository Resource

Creates a local gems repository.

## Example Usage

```hcl
resource "artifactory_local_gems_repository" "terraform-local-test-gems-repo" {
  key = "terraform-local-test-gems-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Gems repository type closely matches with arguments for Generic repository type.
