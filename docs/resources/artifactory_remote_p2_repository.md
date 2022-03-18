# Artifactory Remote P2 Repository Resource

Creates a remote P2 repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/P2+Repositories)


## Example Usage
To create a new Artifactory remote P2 repository called my-remote-p2.

```hcl
resource "artifactory_remote_p2_repository" "my-remote-p2" {
  key                         = "my-remote-p2"
  url                         = "http://testartifactory.io/artifactory/example-p2/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote P2 repository type closely match with arguments for remote Generic repository type.