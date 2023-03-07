---
subcategory: "Remote Repositories"
---
# Artifactory Remote Cargo Repository Resource

Creates a remote Cargo repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Cargo+Registry).


## Example Usage

```hcl
resource "artifactory_remote_cargo_repository" "my-remote-cargo" {
  key                 = "my-remote-cargo"
  anonymous_access    = true
  enable_sparse_index = true
  url                 = "https://github.com/rust-lang/crates.io-index"
  git_registry_url    = "https://github.com/rust-lang/foo.index"
}
```
## Note
If you get a 400 error: `"Custom Base URL should be defined prior to creating a Cargo repository"`,
you must set the base url at: `http://${host}/ui/admin/configuration/general`

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `anonymous_access` - (Required) Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is `false`.
* `enable_sparse_index` - (Optional) Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is `false`.
* `git_registry_url` - (Optional) This is the index url, expected to be a git repository. Default value is `https://github.com/rust-lang/crates.io-index`.


## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_cargo_repository.my-remote-cargo my-remote-cargo
```
