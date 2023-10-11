---
subcategory: "Remote Repositories"
---
# Artifactory Remote Hugging Face Repository Resource

Provides a remote Hugging Face repository. 

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/set-up-remote-hugging-face-repositories).

## Example Usage

```hcl
resource "artifactory_remote_huggingfaceml_repository" "huggingfaceml-remote" {
  key = "huggingfaceml-remote-foo25"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).

The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_huggingfaceml_repository.huggingfaceml-remote huggingfaceml-remote
```
