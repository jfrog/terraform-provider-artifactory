---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Resource

Creates remote Docker repository resource. 

Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Docker+Registry)


## Example Usage

```hcl
resource "artifactory_remote_docker_repository" "my-remote-docker" {
  key                            = "my-remote-docker"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/registry-1.docker.io/**"]
  enable_token_authentication    = true
  url                            = "https://registry-1.docker.io/"
  block_pushing_schema1          = true
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
* `block_pushing_schema1` - (Optional) When set, Artifactory will block the pulling of Docker images with manifest v2
  schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1
  that exist in the cache.
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI.
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. Default to `["**"]`
* `curated` - (Optional, Default: `false`) Enable repository to be protected by the Curation service.
* `project_id` (Optional) Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id will default to their default project as you have set up in your account.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_docker_repository.my-remote-docker my-remote-docker
```
