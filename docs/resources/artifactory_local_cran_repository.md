# Artifactory Local Cran Repository Resource

Creates a local cran repository.

## Example Usage

```hcl
resource "artifactory_local_cran_repository" "terraform-local-test-cran-repo" {
  key = "terraform-local-test-cran-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Cran repository type closely matches with arguments for Generic repository type.
