# Artifactory Local Pub Repository Resource

Creates a local pub repository.

## Example Usage

```hcl
resource "artifactory_local_pub_repository" "terraform-local-test-pub-repo" {
  key = "terraform-local-test-pub-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Pub repository type closely match with arguments for Generic repository type.
