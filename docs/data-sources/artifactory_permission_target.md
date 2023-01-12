# Artifactory Permission Target Data Source

Provides an Artifactory permission target data source. This can be used to read the configuration of permission targets in artifactory.

## Example Usage

```hcl
#
data "artifactory_permission_target" "target1" {
  name  = "my_permission"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the permission target.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `repo` - Repository permission configuration.
  * `includes_pattern` - Pattern of artifacts to include.
  * `excludes_pattern` - Pattern of artifacts to exclude.
  * `repositories` - List of repositories this permission target is applicable for. You can specify the
    name `ANY` in the repositories section in order to apply to all repositories, `ANY REMOTE` for all remote
    repositories and `ANY LOCAL` for all local repositories. The default value will be `[]` if nothing is specified.
  * `actions` -
    * `users` - Users this permission target applies for.
    * `groups` - Groups this permission applies for.
* `build` - Same as repo but for artifactory-build-info permissions.
* `release_bundle` - Same as repo but for release-bundles permissions.
