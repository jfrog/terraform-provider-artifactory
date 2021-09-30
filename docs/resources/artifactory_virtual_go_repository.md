# Artifactory Virtual Repository Resource

Provides an Artifactory virtual repository resource, but with specific go lang features. This should be preferred over the original
one-size-fits-all `artifactory_virtual_repository`. 

## Example Usage

```hcl
resource "artifactory_virtual_go_repository" "baz-go" {
  key          = "baz-go"
  package_type = "go"
  repo_layout_ref = "go-default"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**/github.com/**",
    "**/go.googlesource.com/**"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `package_type` - (Required)
* `repositories` - (Required, but may be empty)
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `artifactory_requests_can_retrieve_remote_artifacts` - (Optional)
* `key_pair` - (Optional) - Key pair to use for... well, I'm not sure. Maybe ssh auth to remote repo?
* `default_deployment_repo` - (Optional)
* `external_dependencies_enabled` - (Optional). Shorthand for "Enable 'go-import' Meta Tags" on the UI. This must be set to true in order to use the allow list
* `external_dependencies_patterns` - (Optional) - 'go-import' Allow List on the UI.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_repository.foo foo
```
