# Artifactory FileInfo Data Source

Provides an Artifactory FileInfo datasource. This can be used to read metadata of files stored in Artifactory repositories.

## Example Usage

```hcl
data "artifactory_fileinfo" "my-file" {
   repository = "repo-key"
   path       = "/path/to/the/artifact.zip" 
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) Name of the repository where the file is stored.

* `path` - (Required) The path to the file within the repository.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `created` - The timestamp for when the file was created.

* `created_by` - The Artifactory user who created the file.

* `last_modified` - The timestamp when the file was last modified.

* `modified_by` - The Artifactory user who last modified the file.

* `last_updated` - The timestamp when the file was last updated.

* `mimetype` - The mimetype of the file.

* `size` - The size of the file, in bytes.

* `download_uri` - The URI that can be used to download the file.

* `md5` - MD5 checksum of the file.

* `sha1` - SHA1 checksum of the file.

* `sha256` - SHA256 checksum of the file.
