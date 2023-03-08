---
subcategory: "Local Repositories"
---

# Artifactory Local RPM Repository Data Source

Retrieves a local RPM repository.

## Example Usage

```hcl
data "artifactory_local_rpm_repository" "local-test-rpm-repo-basic" {
  key = "local-test-rpm-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `yum_root_depth` - The depth, relative to the repository's root folder, where RPM metadata is created. This
  is useful when your repository contains multiple RPM repositories under parallel hierarchies. For example, if your
  RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4. Once the number of snapshots exceeds
  this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique snapshots are
  not cleaned up.
* `calculate_yum_metadata` - Default: `false`.
* `enable_file_lists_indexing` - Default: `false`.
* `yum_group_file_names` - A comma separated list of XML file names containing RPM group component
  definitions. Artifactory includes the group definitions as part of the calculated RPM metadata, as well as
  automatically generating a gzipped version of the group files, if required. Default is empty string.
* `primary_keypair_ref` - The primary GPG key to be used to sign packages.
* `secondary_keypair_ref` - The secondary GPG key to be used to sign packages.
