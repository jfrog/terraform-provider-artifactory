# Artifactory Local Go Repository Resource

Creates a local go repository. 

## Example Usage

```hcl
resource "artifactory_local_go_repository" "terraform-local-test-go-repo" {
  key                 = "terraform-local-test-go-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Go repository type closely matches with arguments for Generic repository type. 