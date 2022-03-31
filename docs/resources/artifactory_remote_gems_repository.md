# Artifactory Remote Gems Repository Resource

Creates a remote Gems repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/RubyGems+Repositories)


## Example Usage
To create a new Artifactory remote Gems repository called my-remote-gems.

```hcl
resource "artifactory_remote_gems_repository" "my-remote-gems" {
  key                         = "my-remote-gems"
  url                         = "https://rubygems.org/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Gems repository type closely match with arguments for remote Generic repository type.