# Artifactory Local Sbt Repository Resource

Creates a local sbt repository.

## Example Usage

```hcl
resource "artifactory_local_sbt_repository" "terraform-local-test-sbt-repo" {
  key = "terraform-local-test-sbt-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Sbt repository type closely match with arguments for Generic repository type.
