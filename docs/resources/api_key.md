---
subcategory: "Deprecated"
---
# Artifactory API Key Resource

Provides an Artifactory API key resource. This can be used to create and manage Artifactory API keys.

~> **Note:** API keys will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

!> As notified in [Artifactory 7.47.10](https://jfrog.com/help/r/jfrog-release-information/artifactory-7.47.10-cloud-self-hosted), support for API Key is slated to be removed in a future release. To ease customer migration to [reference tokens](https://jfrog.com/help/r/jfrog-platform-administration-documentation/user-profile), which replaces API key, we are disabling the ability to create new API keys at the end of Q3 2024. The ability to use API keys will be removed at the end of Q4 2024. For more information, see [JFrog API Key Deprecation Process](https://jfrog.com/help/r/jfrog-platform-administration-documentation/jfrog-api-key-deprecation-process).

## Example Usage

```hcl
# Create a new Artifactory API key for the configured user
resource "artifactory_api_key" "ci" {}
```

## Attribute Reference

The following attributes are exported:

* `api_key` - The API key. Deprecated.

## Import

A user's API key can be imported using any identifier, e.g.

```
$ terraform import artifactory_api_key.test import
```
