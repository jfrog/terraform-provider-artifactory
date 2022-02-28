# Artifactory Local RPM Repository Resource

Creates a local RPM repository

## Example Usage

```hcl
resource "artifactory_local_rpm_repository" "terraform-local-test-rpm-repo-basic" {
  key                        = "terraform-local-test-rpm-repo-basic"
  yum_root_depth             = 5
  calculate_yum_metadata     = true
  enable_file_lists_indexing = true
  yum_group_file_names       = "file-1.xml,file-2.xml"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `yum_root_depth` - (Optional) - The depth, relative to the repository's root folder, where RPM metadata is created. This is useful when your repository contains multiple RPM repositories under parallel hierarchies. For example, if your RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4. Once the number of snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.
* `calculate_yum_metadata` - (Optional)
* `enable_file_lists_indexing` - (Optional)
* `yum_group_file_names` - (Optional) - A list of XML file names containing RPM group component definitions. Artifactory includes the group definitions as part of the calculated RPM metadata, as well as automatically generating a gzipped version of the group files, if required.

Arguments for RPM repository type closely match with arguments for Generic repository type.
