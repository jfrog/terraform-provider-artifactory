# Artifactory Local Conda Repository Resource

Creates a local conda repository.

## Example Usage

```hcl
resource "artifactory_local_conda_repository" "terraform-local-test-conda-repo" {
  key = "terraform-local-test-conda-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Conda repository type closely match with arguments for Generic repository type.
