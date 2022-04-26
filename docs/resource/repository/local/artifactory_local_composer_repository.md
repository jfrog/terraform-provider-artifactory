# Artifactory Local Php-Composer Repository Resource

Creates a local composer repository.

## Example Usage

```hcl
resource "artifactory_local_composer_repository" "terraform-local-test-composer-repo" {
  key = "terraform-local-test-composer-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) - the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Composer repository type closely match with arguments for Generic repository type.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_composer_repository.terraform-local-test-composer-repo terraform-local-test-composer-repo
```