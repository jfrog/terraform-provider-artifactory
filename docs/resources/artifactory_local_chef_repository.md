# Artifactory Local Chef Repository Resource

Creates a local chef repository. 

## Example Usage

```hcl
resource "artifactory_local_chef_repository" "terraform-local-test-chef-repo" {
  key                 = "terraform-local-test-chef-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Chef repository type closely matches with arguments for Generic repository type. 