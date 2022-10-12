---
subcategory: "Security"
---
# Artifactory API Key Resource

Provides an Artifactory API key resource. This can be used to create and manage Artifactory API keys.

~> **Note:** API keys will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).


## Example Usage

```hcl
# Create a new Artifactory API key for the configured user
resource "artifactory_api_key" "ci" {}
```

## Attribute Reference

The following attributes are exported:

* `api_key` - The API key. Deprecated. An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform).
  In September 2022, the option to block the usage/creation of API Keys will be enabled by default, with the option for admins to change it back to enable API Keys.
  In January 2023, API Keys will be deprecated all together and the option to use them will no longer be available.
  It is recommended to use scoped tokens instead - `artifactory_scoped_token` resource.
  Please check the [release notes](https://www.jfrog.com/confluence/display/JFROG/Artifactory+Release+Notes#ArtifactoryReleaseNotes-Artifactory7.38.4).

## Import

A user's API key can be imported using any identifier, e.g.

```
$ terraform import artifactory_api_key.test import
```
