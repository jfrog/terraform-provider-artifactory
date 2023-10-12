---
subcategory: "Local Repositories"
---
# Artifactory Local Hugging Face Repository Resource

Creates a local Hugging Face repository.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/set-up-local-hugging-face-repositories).

## Example Usage

```hcl
resource "artifactory_local_huggingfaceml_repository" "local-huggingfaceml-repo" {
  key = "local-huggingfaceml-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).

The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_huggingfaceml_repository.local-huggingfaceml-repo local-huggingfaceml-repo
```
