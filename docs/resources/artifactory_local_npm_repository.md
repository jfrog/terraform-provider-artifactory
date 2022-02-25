# Artifactory Local NPM Repository Resource

Creates a local npm repository.

## Example Usage

```hcl
resource "artifactory_local_npm_repository" "terraform-local-test-npm-repo" {
  key = "terraform-local-test-npm-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for NPM repository type closely match with arguments for Generic repository type.
