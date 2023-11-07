---
subcategory: "Remote Repositories"
---
# Artifactory Remote npm Repository Resource

Creates a remote Npm repository. 
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/npm+Registry).


## Example Usage

```hcl
resource "artifactory_remote_npm_repository" "npm-remote" {
  key                                  = "npm-remote"
  url                                  = "https://registry.npmjs.org"
  list_remote_folder_items             = true
  mismatching_mime_types_override_list = "application/json,application/xml"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `curated` - (Optional, Default: `false`) Enable repository to be protected by the Curation service.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_npm_repository.npm-remote npm-remote
```
