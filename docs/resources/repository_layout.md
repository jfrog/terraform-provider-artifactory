---
subcategory: "Configuration"
---
# Artifactory Repository Layout Resource

This resource can be used to manage Artifactory's Repository Layout settings. See [Repository Layouts](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts) in the Artifactory Wiki documentation for more details.

~>The `artifactory_repository_layout` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
resource "artifactory_repository_layout" "custom-layout" {
  name                                = "custom-layout"
  artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
  distinctive_descriptor_path_pattern = true
  descriptor_path_pattern             = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).pom"
  folder_integration_revision_regexp  = "Foo"
  file_integration_revision_regexp    = "Foo|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)
* `artifact_path_pattern` - (Required) Please refer to: [Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-ModulesandPathPatternsusedbyRepositoryLayouts) in the Artifactory Wiki documentation.
* `distinctive_descriptor_path_pattern` - (Optional) When set, `descriptor_path_pattern` will be used. Default to `false`.
* `descriptor_path_pattern` - (Optional) Please refer to: [Descriptor Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-DescriptorPathPatterns) in the Artifactory Wiki documentation.
* `folder_integration_revision_regexp` - (Optional) A regular expression matching the integration revision string appearing in a folder name as part of the artifact's path. For example, `SNAPSHOT`, in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use `.*`
* `file_integration_revision_regexp` - (Optional) A regular expression matching the integration revision string appearing in a file name as part of the artifact's path. For example, `SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))`, in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use `.*`

## Import

Repository layout can be imported using its name, e.g.

```
$ terraform import artifactory_repository_layout.custom-layout custom-layout
```
