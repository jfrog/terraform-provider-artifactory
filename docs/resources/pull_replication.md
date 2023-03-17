---
subcategory: "Replication"
---
# Artifactory Pull Replication Resource

Provides an Artifactory pull replication resource. This can be used to create and manage pull replication in Artifactory
for a local or remote repo. Pull replication provides a convenient way to proactively populate a remote cache, and is very useful 
when waiting for new artifacts to arrive on demand (when first requested) is not desirable due to network latency.
See the [Official Documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PullReplication).

## Example Usage

```hcl
# Create a replication between two artifactory local repositories
resource "artifactory_local_maven_repository" "provider_test_source" {
	key          = "provider_test_source"
}

resource "artifactory_remote_maven_repository" "provider_test_dest" {
	key          = "provider_test_dest"
	url          = "https://example.com/artifactory/${artifactory_local_maven_repository.artifactory_local_maven_repository.key}"
	username     = "foo"
	password     = "bar"
}

resource "artifactory_pull_replication" "remote-rep" {
	repo_key                 = "${artifactory_remote_maven_repository.provider_test_dest.key}"
	cron_exp                 = "0 0 * * * ?"
	enable_event_replication = true
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required) Repository name.
* `cron_exp` - (Required) A valid CRON expression that you can use to control replication frequency. Eg: `0 0 12 * * ? *`, `0 0 2 ? * MON-SAT *`. Note: use 6 or 7 parts format - Seconds, Minutes Hours, Day Of Month, Month, Day Of Week, Year (optional). Specifying both a day-of-week AND a day-of-month parameter is not supported. One of them should be replaced by `?`. Incorrect: `* 5,7,9 14/2 * * WED,SAT *`, correct: `* 5,7,9 14/2 ? * WED,SAT *`. See details in [Cron Trigger Tutorial](http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html).
* `enable_event_replication` - (Optional) When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. added, deleted or property change.
* `url` - (Optional) The URL of the target local repository on a remote Artifactory server. For some package types, you need to prefix the repository key in the URL with api/<pkg>. 
   For a list of package types where this is required, see the [note](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-anchorPREFIX). 
   Required for local repository, but not needed for remote repository.
* `username` - (Optional) Required for local repository, but not needed for remote repository.
* `password` - (Optional) Required for local repository, but not needed for remote repository.
* `enabled` - (Optional) When set, this replication will be enabled when saved.
* `sync_deletes` - (Optional) When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata).
* `sync_properties` - (Optional) When set, the task also synchronizes the properties of replicated artifacts.
* `sync_statistics` - (Optional) When set, artifact download statistics will also be replicated. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery.
* `path_prefix` - (Optional) Only artifacts that located in path that matches the subpath within the remote repository will be replicated.
* `check_binary_existence_in_filestore` - (Optional) When true, enables distributed checksum storage. For more information, see
  [Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).

## Import

Pull replication config can be imported using its repo key, e.g.

```
$ terraform import artifactory_pull_replication.foo-rep repository-key
```
