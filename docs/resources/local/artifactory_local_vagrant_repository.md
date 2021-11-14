# Artifactory Local Vagrant Repository Resource

Creates a local vagrant repository. 

## Example Usage

```hcl
resource "artifactory_local_vagrant_repository" "terraform-local-test-vagrant-repo" {
  key                 = "terraform-local-test-vagrant-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Vagrant repository type closely matches with arguments for Generic repository type. 