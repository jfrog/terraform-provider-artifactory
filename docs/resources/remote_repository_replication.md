---
subcategory: "Replication"
---
# Artifactory Remote Repository Replication Resource

Provides a remote repository replication resource, also referred to as Artifactory pull replication. 
This resource provides a convenient way to proactively populate a remote cache, and is very useful when waiting for new artifacts to arrive on demand (when first requested) is not desirable due to network latency. See [official documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PullReplication).


## Example Usage

```hcl
variable "artifactory_url" {
  description = "The base URL of the Artifactory deployment"
  type        = string
}

resource "artifactory_local_maven_repository" "provider_test_source" {
  key          = "provider_test_source"
}

resource "artifactory_remote_maven_repository" "provider_test_dest" {
  key          = "provider_test_dest"
  url          = "${var.artifactory_url}/artifactory/${artifactory_local_maven_repository.artifactory_local_maven_repository.key}"
  username     = "foo"
  password     = "bar"
}

resource "artifactory_remote_repository_replication" "remote-rep" {
  repo_key                            = "${artifactory_remote_maven_repository.provider_test_dest.key}"
  cron_exp                            = "0 0 * * * ?"
  enable_event_replication            = true
  enabled                             = true
  sync_deletes                        = false
  sync_properties                     = true
  include_path_prefix_pattern         = "/some-repo/"
  exclude_path_prefix_pattern         = "/some-other-repo/"
  check_binary_existence_in_filestore = false
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required) Repository name.
* `cron_exp` - (Required) A valid CRON expression that you can use to control replication frequency. Eg: `0 0 12 * * ? *`, `0 0 2 ? * MON-SAT *`. Note: use 6 or 7 parts format - Seconds, Minutes Hours, Day Of Month, Month, Day Of Week, Year (optional). Specifying both a day-of-week AND a day-of-month parameter is not supported. One of them should be replaced by `?`. Incorrect: `* 5,7,9 14/2 * * WED,SAT *`, correct: `* 5,7,9 14/2 ? * WED,SAT *`. See details in [Cron Trigger Tutorial](https://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html).
* `enable_event_replication` - (Optional) When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.
com/confluence/display/JFROG/User+Profile#UserProfile-IdentityTokenidentitytoken).
* `sync_deletes` - (Optional) When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.
* `sync_properties` - (Optional) When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`.
* `enabled` - (Optional) When set, enables replication of this repository to the target specified in `url` attribute. Default value is `true`.
* `include_path_prefix_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form of `x/y/**/z/*`. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included `(**/*)`.
* `exclude_path_prefix_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form of `x/y/**/z/*`. By default, no artifacts are excluded.
* `replication_key` - (Computed) Replication ID, the value is unknown until the resource is created. Can't be set or updated.
* `check_binary_existence_in_filestore` - (Optional) Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise Plus license. When true, enables distributed checksum storage. For more information, see [Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).

## Import

Push replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_remote_repository_replication.foo-rep provider_test_source
```
