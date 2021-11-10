# Artifactory Local npm Repository Resource

Creates a local npm repository and allows for the creation of a 

## Example Usage

```hcl
resource "artifactory_local_npm_repository" "terraform-local-test-npm-repo-basic" {
  key                 = "terraform-local-test-npm-repo-basic"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo