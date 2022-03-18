# Artifactory Remote RPM Repository Resource

Creates a remote RPM repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/RPM+Repositories)


## Example Usage
To create a new Artifactory remote RPM repository called my-remote-rpm.

```hcl
resource "artifactory_remote_rpm_repository" "my-remote-rpm" {
  key                         = "my-remote-rpm"
  url                         = "http://mirror.centos.org/centos/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote RPM repository type closely match with arguments for remote Generic repository type.