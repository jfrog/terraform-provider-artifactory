# Artifactory Remote Conan Repository Resource

Creates a remote Conan repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Conan+Repositories)


## Example Usage
To create a new Artifactory remote Conan repository called my-remote-conan.

```hcl
resource "artifactory_remote_conan_repository" "my-remote-conan" {
  key                         = "my-remote-conan"
  url                         = "https://conan.bintray.com"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Conan repository type closely match with arguments for remote Generic repository type.