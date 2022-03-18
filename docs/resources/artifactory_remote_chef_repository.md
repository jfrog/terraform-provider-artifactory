# Artifactory Remote Chef Repository Resource

Creates a remote Chef repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Chef+Cookbook+Repositories)


## Example Usage
To create a new Artifactory remote Chef repository called my-remote-chef.

```hcl
resource "artifactory_remote_chef_repository" "my-remote-chef" {
  key                         = "my-remote-chef"
  url                         = "https://supermarket.chef.io"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Chef repository type closely match with arguments for remote Generic repository type.