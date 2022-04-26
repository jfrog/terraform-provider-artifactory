# Artifactory Local Ivy Repository Resource

Creates a local ivy repository.

## Example Usage

```hcl
resource "artifactory_local_ivy_repository" "terraform-local-test-ivy-repo" {
  key = "terraform-local-test-ivy-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) - the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Ivy repository type closely match with arguments for Gradle repository type.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_ivy_repository.terraform-local-test-ivy-repo terraform-local-test-ivy-repo
```