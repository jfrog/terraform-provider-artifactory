# Artifactory Local Ivy Repository Resource

Creates a local ivy repository.

## Example Usage

```hcl
resource "artifactory_local_ivy_repository" "terraform-local-test-ivy-repo" {
  key = "terraform-local-test-ivy-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Ivy repository type closely matches with arguments for Generic repository type.
