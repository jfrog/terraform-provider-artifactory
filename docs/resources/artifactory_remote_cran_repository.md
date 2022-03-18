# Artifactory Remote Cran Repository Resource

Creates a remote Cran repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/CRAN+Repositories)


## Example Usage
To create a new Artifactory remote Cran repository called my-remote-cran.

```hcl
resource "artifactory_remote_cran_repository" "my-remote-cran" {
  key                         = "my-remote-cran"
  url                         = "https://cran.r-project.org/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Cran repository type closely match with arguments for remote Generic repository type.