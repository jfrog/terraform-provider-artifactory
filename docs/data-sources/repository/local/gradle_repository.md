---
subcategory: "Local Repositories"
---

# Artifactory Local Gradle Repository Data Source

Retrieves a local Gradle repository.

## Example Usage

```hcl
data "artifactory_local_gradle_repository" "local-test-gradle-repo-basic" {
  key = "local-test-gradle-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `checksum_policy_type` - Checksum policy determines how Artifactory behaves when a client checksum for a
  deployed resource is missing or conflicts with the locally calculated checksum (bad checksum). The options are
  `client-checksums` and `generated-checksums`. For more details, please refer
  to [Checksum Policy](https://www.jfrog.com/confluence/display/JFROG/Local+Repositories#LocalRepositories-ChecksumPolicy)
  .
* `snapshot_version_behavior` - Specifies the naming convention for Maven SNAPSHOT versions. The options are
  -
  * `unique`: Version number is based on a time-stamp (default)
  * `non-unique`: Version number uses a self-overriding naming pattern of artifactId-version-SNAPSHOT.type
  * `deployer`: Respects the settings in the Maven client that is deploying the artifact.
* `max_unique_snapshots` - The maximum number of unique snapshots of a single artifact to store. Once the
  number of snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no
  limit, and unique snapshots are not cleaned up.
* `handle_releases` - If set, Artifactory allows you to deploy release artifacts into this repository.
  Default is `true`.
* `handle_snapshots` - If set, Artifactory allows you to deploy snapshot artifacts into this repository.
  Default is `true`.
* `suppress_pom_consistency_checks` - By default, Artifactory keeps your repositories healthy by refusing
  POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match
  the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by
  setting the Suppress POM Consistency Checks checkbox. True by default for Gradle repository.
