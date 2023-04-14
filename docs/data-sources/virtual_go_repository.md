---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Go Repository Data Source

Retrieves a virtual Go repository.

## Example Usage

```hcl
data "artifactory_virtual_go_repository" "virtual-go" {
  key = "virtual-go"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `external_dependencies_enabled` - (Optional) Shorthand for "Enable 'go-import' Meta Tags" on the UI. This must be set to true in order to use the allow list. 
  When checked (default), Artifactory will automatically follow remote VCS roots in 'go-import' meta tags to download remote modules.
* `external_dependencies_patterns` - (Optional) 'go-import' Allow List on the UI.
