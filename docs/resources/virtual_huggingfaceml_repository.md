---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Hugging Face ML Repository Resource

Creates a virtual Hugging Face ML repository that aggregates local and remote Hugging Face ML repositories.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/set-up-virtual-hugging-face-repositories).

## Example Usage

```hcl
resource "artifactory_local_huggingfaceml_repository" "huggingfaceml-local" {
  key = "huggingfaceml-local"
}

resource "artifactory_remote_huggingfaceml_repository" "huggingfaceml-remote" {
  key = "huggingfaceml-remote"
}

resource "artifactory_virtual_huggingfaceml_repository" "my-virtual-huggingfaceml" {
  key = "my-virtual-huggingfaceml"
  repositories = [
    artifactory_local_huggingfaceml_repository.huggingfaceml-local.key,
    artifactory_remote_huggingfaceml_repository.huggingfaceml-remote.key
  ]
  description = "Virtual Hugging Face ML repository aggregating local and remote"
  notes       = "Internal repository"
  depends_on = [
    artifactory_local_huggingfaceml_repository.huggingfaceml-local,
    artifactory_remote_huggingfaceml_repository.huggingfaceml-remote
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `repo_layout_ref` - (Optional, Computed) Sets the layout that the repository should use for storing and identifying modules. Defaults to `simple-default` for Hugging Face ML repositories.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_huggingfaceml_repository.my-virtual-huggingfaceml my-virtual-huggingfaceml
```
