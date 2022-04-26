# Artifactory Remote Pub Repository Resource

Creates a remote Pub repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Pub+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_pub_repository" "my-remote-pub" {
  key                         = "my-remote-pub"
  url                         = "https://pub.dartlang.org"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) - A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repository URL.

Arguments for remote Pub repository type closely match with arguments for remote Generic repository type.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_pub_repository.my-remote-pub my-remote-pub
```