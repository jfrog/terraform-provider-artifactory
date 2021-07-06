# Xray Watch Resource

Provides an Xray watch resource. This can be used to create and manage Xray watches.

## Example Usage

### Example Watch for all repositories
```hcl
resource "artifactory_xray_watch" "all_repos" {
  name        = "all-repositories"
  description = "all repositories"

  all_repositories {
    package_types = ["NuGet", "Docker"]
    paths         = ["path/*"]
    names         = ["name2", "name1"]
    mime_types    = ["application/zip"]

    property {
      key   = "field4"
      value = "value 4"
    }
    property {
      key   = "field2"
      value = "value 2"
    }
  }
	
  repository_paths {
    include_patterns = [
      "path1/**",
      "another-path/**",
    ]

    exclude_patterns = [
      "path1/ignore/**",
      "another-path/ignore**",
    ]
  }
}
```

### Example Watch for named repositories
```hcl
resource "artifactory_local_repository" "example" {
  key 	       = "local-repo"
  package_type = "generic"
  xray_index   = true
}

resource "artifactory_xray_watch" "watch" {
  name        = "named-local-repo"
  description = "all repositories"

  repository {
    name = artifactory_local_repository.example.key

    package_types = ["Generic"]
    paths         = ["path/*"]
    mime_types    = ["application/zip"]

    property {
      key   = "field1"
      value = "value 1"
    }
    property {
      key   = "field2"
      value = "value 2"
    }
  }

  repository_paths {
    include_patterns = [
      "path1/**",
    ]
    exclude_patterns = [
      "path1/ignore/**",
    ]
  }
}
```

### Example Watch for all builds
```hcl
resource "artifactory_xray_watch" "all_builds" {
  name        = "all_builds"
  description = "all builds"
  active      = true

  all_builds {
    bin_mgr_id = "default"
  }

  policy {
    name = "policy-1"
    type = "security"
  }
}
```

### Example Watch for all builds, filtered by pattern
```hcl
resource "artifactory_xray_watch" "pattern_builds" {
  name        = "pattern_builds"
  description = "builds by pattern"
  active      = true

  all_builds {
    include_patterns = ["hello/**", "apache/**"]
    exclude_patterns = ["apache/bad*", "world/**"]
    bin_mgr_id       = "default"        
  }

  policy {
    name = "policy-1"
    type = "security"
  }
}
```

### Example Watch for named builds
```hcl
resource "artifactory_xray_watch" "named_builds" {
  name        = "named_builds"
  description = "named builds"
  active      = true

  build {
    name       = "build1"
    bin_mgr_id = "default"
  }

  policy {
    name = "policy-1"
    type = "security"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the watch. Changing this forces a new resource to be created.
* `description` - (Optional) The description of the watch.
* `active` - (Optional) Is the watch active? Only can be true when a watch has repository or builds defined, and at least 1 policy.
* `all_repositories` - (Optional) This block defines the configuration for all repositores. It conflicts with the `repository` block.
  * `names` - (Optional) The list of repositories to filter.
  * `bin_mgr_id` - (Optional) The Binary Manager ID. Defaults to `default`.
  * `package_types` - (Optional) A list of package types to filter.
  * `paths` - (Optional) A list of paths to filter.
  * `mime_types` - (Optional) A list of mime types to filter.
  * `property` - (Optional) This block defines a property to filter. You can specify multiple blocks for multiple properties.
    * `key` - (Required) The property's key.
    * `value` - (Required) The property's value.
* `repository` - (Optional) This block defines the configuration for a repository.  You can specify multiple blocks for multiple repositories. It conflicts with the `all_repositories` block.
  * `name` - (Required) The name of the repository.
  * `bin_mgr_id` - (Optional) The Binary Manager ID. Defaults to `default`.
  * `package_types` - (Optional) A list of package types to filter.
  * `paths` - (Optional) A list of paths to filter.
  * `mime_types` - (Optional) A list of mime types to filter.
  * `property` - (Optional) This block defines a property to filter. You can specify multiple blocks for multiple properties.
    * `key` - (Required) The property's key.
    * `value` - (Required) The property's value.
* `repository_paths` - (Optional) This block defines the paths for filtering any repository.
  * `include_patterns` - (Optional) A list of strings, which will be included. Simple comma separated wildcard patterns for repository artifact paths (with no leading slash). Ant-style path expressions are supported (*, \*\*, ?). Example: "org/apache/\*\*".
  * `exclude_patterns` - (Optional) A list of strings, which will be excluded. Simple comma separated wildcard patterns for repository artifact paths (with no leading slash). Ant-style path expressions are supported (*, \*\*, ?). Example: "org/apache/\*\*".
* `all_builds` - (Optional) This block defines the configuration for all builds. It conflicts with the `build` block.
  * `bin_mgr_id` - (Optional) The Binary Manager ID. Defaults to `default`.
  * `include_patterns` - (Optional) A list of strings in an ant-style wildcard format to specify build names that will be included.
  * `exclude_patterns` - (Optional) A list of strings in an ant-style wildcard format to specify build names that will be excluded.
* `build` - (Optional) This block defines the configuration for a build. It conflicts with the `all_build` block.
  * `name` - (Required) The name of the build.
  * `bin_mgr_id` - (Optional) The Binary Manager ID. Defaults to `default`.
* `policy` - (Optional) This block defines the policy associated with the watch. You can specify multiple blocks for multiple policies.
  * `name` - (Required) The name of the policy.
  * `type` - (Required) The type of the policy. Allowed values: `security` or `license`.

## References
- https://www.jfrog.com/confluence/display/JFROG/Configuring+Xray+Watches
- https://www.jfrog.com/confluence/display/JFROG/Xray+REST+API#XrayRESTAPI-WATCHES

## Import

A watch can be imported using their name, e.g.

```
$ terraform import artifactory_xray_watch.foo foo
```
