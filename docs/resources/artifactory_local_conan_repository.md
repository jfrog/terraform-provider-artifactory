# Artifactory Local Conan Repository Resource

Creates a local conan repository.

## Example Usage

```hcl
resource "artifactory_local_conan_repository" "terraform-local-test-conan-repo" {
  key = "terraform-local-test-conan-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Conan repository type closely matches with arguments for Generic repository type.
