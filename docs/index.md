# Artifactory Provider

The [Artifactory](https://jfrog.com/artifactory/) provider is used to interact with the resources supported by Artifactory. The provider needs to be configured with the proper credentials before it can be used.

Links to documentation for specific resources can be found in the table of contents to the left.

This provider requires access to Artifactory APIs, which are only available in the _licensed_ pro and enterprise editions. You can determine which license you have by accessing the following the URL `${host}/artifactory/api/system/licenses/`.

You can either access it via API, or web browser - it require admin level credentials.

```sh
curl -sL ${host}/artifactory/api/system/licenses/ | jq .
{
  "type" : "Enterprise Plus Trial",
  "validThrough" : "Jan 29, 2022",
  "licensedTo" : "JFrog Ltd"
}
```

## Terraform CLI version support

Current version support [Terraform Protocol v5](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-5) which mean Terraform CLI version 0.12 and later. 

~>We will be moving to [Terraform Protocol v6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) in **Q1 2024**. This means only Terraform CLI version 1.0 and later will be supported.

## Example Usage
```tf
# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "9.7.0"
    }
  }
}

# Configure the Artifactory provider
provider "artifactory" {
  url           = "${var.artifactory_url}/artifactory"
  access_token  = "${var.artifactory_access_token}"
}

# Create a new repository
resource "artifactory_local_pypi_repository" "pypi-libs" {
  key             = "pypi-libs"
  repo_layout_ref = "simple-default"
  description     = "A pypi repository for python packages"
}
```

## Authentication
The Artifactory provider supports two ways of authentication. The following methods are supported:
* Access Token
* API Key

### Access Token
Artifactory access tokens may be used via the Authorization header by providing the `access_token` attribute to the provider block. Getting this value from the environment is supported with `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` variables.

Usage:
```tf
# Configure the Artifactory provider
provider "artifactory" {
  url           = "artifactory.site.com/artifactory"
  access_token  = "abc...xy"
}
```

### API Key

!>An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform). In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys. By end of Q1 2024, API Keys will be deprecated all together and the option to use them will no longer be available. See [JFrog API Key Deprecation Process](https://jfrog.com/help/r/jfrog-platform-administration-documentation/jfrog-api-key-deprecation-process).

~>If `access_token` attribute, `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variable is set, the provider will ignore `api_key` attribute.

Artifactory API keys may be used via the `X-JFrog-Art-Api` header by providing the `api_key` attribute in the provider block.

Usage:
```tf
# Configure the Artifactory provider
provider "artifactory" {
  url     = "artifactory.site.com/artifactory"
  api_key = "abc...xy"
}
```

## Argument Reference

The following arguments are supported:

* `url` - (Optional) URL of Artifactory. This can also be sourced from the `JFROG_URL` or `ARTIFACTORY_URL` environment variable.
* `access_token` - (Optional) This can also be sourced from `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variables.
* `api_key` - (Optional, deprecated) API key for api auth.
* `check_license` - (Optional) Toggle for pre-flight checking of Artifactory license. Default to `true`.
