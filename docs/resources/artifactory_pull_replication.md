# Artifactory Pull Replication Resource

Provides an Artifactory pull replication resource. This can be used to create and manage pull replication in Artifactory
for a remote repo.

## Example Usage

```hcl
# Create a replication between two artifactory local repositories
resource "artifactory_local_repository" "provider_test_source" {
	key          = "provider_test_source"
	package_type = "maven"
}

resource "artifactory_remote_repository" "provider_test_dest" {
	key          = "provider_test_dest"
	package_type = "maven"
	url          = "https://example.com/artifactory/${artifactory_local_repository.artifactory_local_repository.key}"
	username     = "foo"
	password     = "bar"
}

resource "artifactory_pull_replication" "foo-rep" {
	repo_key                 = "${artifactory_remote_repository.provider_test_dest.key}"
	cron_exp                 = "0 0 * * * ?"
	enable_event_replication = true	
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required)
* `cron_exp` - (Required)
* `enable_event_replication` - (Optional)
* `enabled` - (Optional)
* `sync_deletes` - (Optional)
* `sync_properties` - (Optional)
* `sync_statistics` - (Optional)
* `path_prefix` - (Optional)

## Import

Pull replication config can be imported using its repo key, e.g.

```
$ terraform import artifactory_pull_replication.foo-rep repository-key
```
