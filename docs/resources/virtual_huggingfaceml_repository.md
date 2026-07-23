---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Hugging Face ML Repository Resource

Creates a virtual Hugging Face ML repository that aggregates local and remote Hugging Face ML repositories.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/set-up-virtual-hugging-face-repositories).

~> **Resolution limitation.** Artifactory virtual Hugging Face repositories support resolving models
via `snapshot_download` only. They do not implement symbolic-reference resolution (a request for
revision `main` returns HTTP 400) or the recursive `tree` listing API (returns HTTP 501). As a result,
the modern `huggingface_hub` client (which uses the `tree` API for `snapshot_download` since v1.22.0)
fails against a virtual repo. Working options: use `snapshot_download` on `huggingface_hub` `<= 0.34.4`
(which enumerates via `model_info`), pull single files with an explicit commit SHA
(`hf_hub_download(..., revision="<sha>")`, resolving `main` to a SHA via `model_info` first), or point
clients that need full compatibility (`from_pretrained`, single-file resolve) at the **remote** Hugging
Face repository instead. This is a JFrog platform behavior, not a provider issue. All members must also
use the new Machine Learning repository layout.

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
