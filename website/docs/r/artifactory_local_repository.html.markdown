---
layout: "artifactory"
page_title: "Artifactory: artifactory_local_repository"
sidebar_current: "docs-artifactory-resource-local-repository"
description: |-
  Provides an local repositroy resource.
---

# artifactory_local_repository

Provides an Artifactory local repository resource. This can be used to create and manage Artifactory local repositories.

## Example Usage

```hcl
# Create a new Artifactory local repository called my-local
resource "artifactory_local_repository" "my-local" {
  key          = "my-local"
  package_type = "npm"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `package_type` - (Required)
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `handle_releases` - (Optional) 
* `handle_snapshots` - (Optional) 
* `max_unique_snapshots` - (Optional) 
* `debian_trivial_layout` - (Optional) 
* `checksum_policy_type` - (Optional) 
* `max_unique_tags` - (Optional) 
* `snapshot_version_behavior` - (Optional) 
* `suppress_pom_consistency_checks` - (Optional) 
* `blacked_out` - (Optional) 
* `property_sets` - (Optional) 
* `archive_browsing_enabled` - (Optional) 
* `calculate_yum_metadata` - (Optional) 
* `yum_root_depth` - (Optional) 
* `docker_api_version` - (Optional) 
* `enable_file_lists_indexing` - (Optional) 
* `force_nuget_authentication` - (Optional, Nuget repos only) 

## Import

Local repositories can be imported using their name, e.g.

```
$ terraform import artifactory_local_repository.my-local my-local
```