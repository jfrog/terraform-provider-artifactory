# Artifactory Local Pypi Repository Resource

Creates a local pypi repository.

## Example Usage

```hcl
resource "artifactory_local_pypi_repository" "terraform-local-test-pypi-repo" {
  key = "terraform-local-test-pypi-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) - the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Pypi repository type closely match with arguments for Generic repository type.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_pypi_repository.terraform-local-test-pypi-repo terraform-local-test-pypi-repo
```