# Artifactory Local Puppet Repository Resource

Creates a local puppet repository. 

## Example Usage

```hcl
resource "artifactory_local_puppet_repository" "terraform-local-test-puppet-repo" {
  key                 = "terraform-local-test-puppet-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)

Arguments for Puppet repository type closely matches with arguments for Generic repository type. 