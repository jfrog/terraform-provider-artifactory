# Artifactory Remote Puppet Repository Resource

Creates a remote Puppet repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Puppet+Repositories)


## Example Usage
To create a new Artifactory remote Puppet repository called my-remote-puppet.

```hcl
resource "artifactory_remote_puppet_repository" "my-remote-puppet" {
  key                         = "my-remote-puppet"
  url                         = "https://forgeapi.puppetlabs.com/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Puppet repository type closely match with arguments for remote Generic repository type.