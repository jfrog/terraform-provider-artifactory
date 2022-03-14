# Artifactory File Data Source

Provides an Artifactory file datasource. This can be used to download a file from a given Artifactory repository.

## Example Usage

```hcl
# 
data "artifactory_file" "my-file" {
   repository = "repo-key"
   path = "/path/to/the/artifact.zip"
   output_path = "tmp/artifact.zip"
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) Name of the repository where the file is stored.
* `path` - (Required) The path to the file within the repository.
* `output_path` - (Required) The local path the file should be downloaded to.
* `force_overwrite` - (Optional) If set to true, an existing file in the output_path will be overwritten. Default: false
*  `download_latest_artifact` - (Optional) If set to `true`, HCL will send the artifact path directly to Artifactory without verification 
   if the file exists. Used in maven repositories, where handling releases and snapshots enabled. For example,
   if there is a jar file in the repo, called `3.7-SNAPSHOT/multi1-3.7-20220310.233748-1.jar`, 
   user can download it by setting the path to `3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.jar`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `created` - The time & date when the file was created.
* `created_by` - The user who created the file.
* `last_modified` - The time & date when the file was last modified.
* `modified_by` - The user who last modified the file.
* `last_updated` - The time & date when the file was last updated.
* `mimetype` - The mimetype of the file.
* `size` - The size of the file.
* `download_uri` - The URI that can be used to download the file.
* `md5` - MD5 checksum of the file.
* `sha1` - SHA1 checksum of the file.
* `sha256` - SHA256 checksum of the file.
