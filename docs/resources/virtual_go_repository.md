---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Go Repository Resource

Creates a virtual Go repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Go+Registry#GoRegistry-VirtualRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_go_repository" "baz-go" {
  key                             = "baz-go"
  repo_layout_ref                 = "go-default"
  repositories                    = []
  description                     = "A test virtual repo"
  notes                           = "Internal description"
  includes_pattern                = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                = "com/google/**"
  external_dependencies_enabled   = true
  external_dependencies_patterns  = [
    "**/github.com/**",
    "**/go.googlesource.com/**"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `external_dependencies_enabled` - (Optional) Shorthand for "Enable 'go-import' Meta Tags" on the UI. This must be set to true in order to use the allow list. 
  When checked (default), Artifactory will automatically follow remote VCS roots in 'go-import' meta tags to download remote modules.
* `external_dependencies_patterns` - (Optional) 'go-import' Allow List on the UI.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_go_repository.baz-go baz-go
```
