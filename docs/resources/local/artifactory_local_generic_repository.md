# Artifactory Local Generic Repository Resource

Creates a local generic repository. 

## Example Usage

```hcl
resource "artifactory_local_generic_repository" "terraform-local-test-generic-repo" {
  key                 = "terraform-local-test-generic-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)
