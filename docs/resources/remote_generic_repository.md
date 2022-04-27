---
subcategory: "Remote Repositories"
---
# Artifactory Remote Generic Repository Resource

Creates a remote Generic repository.

## Example Usage

```hcl
resource "artifactory_remote_generic_repository" "my-remote-generic" {
  key                         = "my-remote-generic"
  url                         = "http://testartifactory.io/artifactory/example-generic/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

All generic repo arguments are supported, in addition to:
* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_generic_repository.my-remote-generic my-remote-generic
```
