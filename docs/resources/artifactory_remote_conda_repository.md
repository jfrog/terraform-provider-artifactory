# Artifactory Remote Conda Repository Resource

Creates a remote Conda repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Conda+Repositories)


## Example Usage
To create a new Artifactory remote Conda repository called my-remote-conda.

```hcl
resource "artifactory_remote_conda_repository" "my-remote-conda" {
  key                         = "my-remote-conda"
  url                         = "https://repo.anaconda.com/pkgs/main"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Conda repository type closely match with arguments for remote Generic repository type.