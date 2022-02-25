# Artifactory Local Bower Repository Resource

Creates a local bower repository.

## Example Usage

```hcl
resource "artifactory_local_bower_repository" "terraform-local-test-bower-repo" {
  key = "terraform-local-test-bower-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Bower repository type closely match with arguments for Generic repository type.
