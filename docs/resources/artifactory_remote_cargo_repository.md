# Artifactory Remote Cargo Repository Resource

Provides an Artifactory remote `cargo` repository resource. This provides cargo specific fields and is the only way to get them
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Cargo+Registry)


## Example Usage
Create a new Artifactory remote cargo repository called my-remote-cargo
for brevity sake, only cargo specific fields are included; for other fields see documentation for
[generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_cargo_repository" "my-remote-cargo" {
  key                 = "my-remote-cargo"
  anonymous_access    = true
  git_registry_url    = "https://github.com/rust-lang/foo.index"
}
```
## Note
If you get a 400 error: `"Custom Base URL should be defined prior to creating a Cargo repository"`, 
you must set the base url at: `http://${host}/ui/admin/configuration/general` 

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `anonymous_access` - (Required) - Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.
* `git_registry_url` - (Optional) - This is the index url, expected to be a git repository. for remote artifactory use "arturl/git/repokey.git"

