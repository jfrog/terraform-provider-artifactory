---
subcategory: "Local Repositories"
---
# Artifactory Local RPM Repository Resource

Creates a local RPM repository.

## Example Usage

```hcl
resource "artifactory_local_rpm_repository" "terraform-local-test-rpm-repo-basic" {
  key                        = "terraform-local-test-rpm-repo-basic"
  yum_root_depth             = 5
  calculate_yum_metadata     = true
  enable_file_lists_indexing = true
  yum_group_file_names       = "file-1.xml,file-2.xml"
  primary_keypair_ref        = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref      = artifactory_keypair.some-keypairGPG2.pair_name
  depends_on                 = [
    artifactory_keypair.some-keypair-gpg-1, 
    artifactory_keypair.some-keypair-gpg-2
  ]
}

resource "artifactory_keypair" "some-keypair-gpg-1" {
  pair_name         = "some-keypair${random_id.randid.id}"
  pair_type         = "GPG"
  alias             = "foo-alias1"
  private_key       = file("samples/gpg.priv")
  public_key        = file("samples/gpg.pub")
  lifecycle {
    ignore_changes  = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_keypair" "some-keypair-gpg-2" {
  pair_name         = "some-keypair${random_id.randid.id}"
  pair_type         = "GPG"
  alias             = "foo-alias2"
  private_key       = file("samples/gpg.priv")
  public_key        = file("samples/gpg.pub")
  lifecycle {
    ignore_changes  = [
      private_key,
      passphrase,
    ]
  }
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `yum_root_depth` - (Optional) The depth, relative to the repository's root folder, where RPM metadata is created. 
This is useful when your repository contains multiple RPM repositories under parallel hierarchies. For example, if 
your RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4. Once the number of snapshots 
exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique 
snapshots are not cleaned up.
* `calculate_yum_metadata` - (Optional) Default: `false`.
* `enable_file_lists_indexing` - (Optional) Default: `false`.
* `yum_group_file_names` - (Optional) A comma separated list of XML file names containing RPM group component definitions. 
Artifactory includes the group definitions as part of the calculated RPM metadata, as well as automatically 
generating a gzipped version of the group files, if required. Default is empty string.
* `primary_keypair_ref` - (Optional) The primary GPG key to be used to sign packages.
* `secondary_keypair_ref` - (Optional) The secondary GPG key to be used to sign packages.



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_rpm_repository.terraform-local-test-rpm-repo-basic terraform-local-test-rpm-repo-basic
```
