# Artifactory Local Gradle Repository Resource

Creates a local Maven repository and allows for the creation of a 

## Example Usage

```hcl
resource "artifactory_local_maven_repository" "terraform-local-test-maven-repo-basic" {
  key                             = "terraform-local-test-maven-repo-basic"
  checksum_policy_type            = "client-checksums"
  snapshot_version_behavior       = "unique"
  max_unique_snapshots            = 10
  handle_releases                 = true
  handle_snapshots                = true
  suppress_pom_consistency_checks = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `checksum_policy_type` - (Optional) - Checksum policy determines how Artifactory behaves when a client checksum for a deployed
  "resource is missing or conflicts with the locally calculated checksum (bad checksum). For more details,
  "please refer to [Checksum Policy](https://www.jfrog.com/confluence/display/JFROG/Local+Repositories#LocalRepositories-ChecksumPolicy)
* `snapshot_version_behavior` - Specifies the naming convention for Maven SNAPSHOT versions.
  The options are -
  Unique: Version number is based on a time-stamp (default)
  Non-unique: Version number uses a self-overriding naming pattern of artifactId-version-SNAPSHOT.type
  Deployer: Respects the settings in the Maven client that is deploying the artifact.
* `max_unique_snapshots` - (Optional) - The maximum number of unique snapshots of a single artifact to store.
  Once the number of snapshots exceeds this setting, older versions are removed.
  A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.
* `handle_releases` - If set, Artifactory allows you to deploy release artifacts into this repository.
* `handle_snapshots` - If set, Artifactory allows you to deploy snapshot artifacts into this repository.
* `suppress_pom_consistency_checks` - (Optional) By default, Artifactory keeps your repositories healthy by refusing POMs with incorrect coordinates (path).
  If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error.
  You can disable this behavior by setting the Suppress POM Consistency Checks checkbox. False by default for Maven repository