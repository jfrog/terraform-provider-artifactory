# Artifactory API Key Resource

Provides an Artifactory API key resource. This can be used to create and manage Artifactory API keys.

~> **Note:** API keys will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```hcl
# Create a new Artifactory API key for the configured user.
resource "artifactory_api_key" "ci" {}
```

## Attribute Reference

The following attributes are exported:

* `api_key` - An API key that can be used with the `X-JFrog-Art-Api` HTTP header to authenticate requests.

## Import

A user's API key can be imported using any identifier.

```
terraform import artifactory_api_key.test import
```
