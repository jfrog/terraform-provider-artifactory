# Artifactory Provider

The [Artifactory](https://jfrog.com/artifactory/) provider is used to interact with the
resources supported by Artifactory. The provider needs to be configured
with the proper credentials before it can be used.

Links to documentation for specific resources can be found in the table of
contents to the left.

This provider requires access to Artifactory APIs, which are only available in the _licensed_ pro and enterprise editions.
You can determine which license you have by accessing the following URL
`${host}/artifactory/api/system/licenses/`

You can either access it via api, or web browser - it does require admin level credentials, but it's one of the few
APIs that will work without a license (side node: you can also install your license here with a `POST`)

```bash
curl -sL ${host}/artifactory/api/system/licenses/ | jq .
{
  "type" : "Enterprise Plus Trial",
  "validThrough" : "Jan 29, 2022",
  "licensedTo" : "JFrog Ltd"
}

```

The following 3 license types (`jq .type`) do **NOT** support APIs:
- Community Edition for C/C++
- JCR Edition
- OSS

## Example Usage
```hcl
# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.17"
    }
  }
}

# Configure the Artifactory provider
provider "artifactory" {
  url = "${var.artifactory_url}/artifactory"
  username = "${var.artifactory_username}"
  password = "${var.artifactory_password}"
}

# Create a new repository
resource "artifactory_local_pypi_repository" "pypi-libs" {
  key             = "pypi-libs"
  repo_layout_ref = "simple-default"
  description     = "A pypi repository for python packages"
}
```

## Authentication
The Artifactory provider supports multiple means of authentication. The following methods are supported:

    * Basic Auth
    * Bearer Token
    * JFrog API Key Header

### Basic Auth
Basic auth may be used by adding a `username` and `password` field to the provider block
Getting this value from the environment is supported with the `ARTIFACTORY_USERNAME` and `ARITFACTORY_PASSWORD` variable

Usage:
```hcl
# Configure the Artifactory provider
provider "artifactory" {
  url = "artifactory.site.com/"
  username = "myusername"
  password = "mypassword"
}
```

### Bearer Token
Artifactory access tokens may be used via the Authorization header by providing the `access_token` field to the provider
block. Getting this value from the environment is supported with the `ARTIFACTORY_ACCESS_TOKEN` variable

Usage:
```hcl
# Configure the Artifactory provider
provider "artifactory" {
  url = "artifactory.site.com/artifactory"
  access_token = "abc...xy"
}
```

### JFrog API Key Header
Artifactory API keys may be used via the `X-JFrog-Art-Api` header by providing the `api_key` field in the provider block.
Getting this value from the environment is supported with the `ARTIFACTORY_API_KEY` variable

Usage:
```hcl
# Configure the Artifactory provider
provider "artifactory" {
  url = "artifactory.site.com/artifactory"
  api_key = "abc...xy"
}
```

## Argument Reference

The following arguments are supported:

* `url` - (Required) URL of Artifactory. This can also be sourced from the `ARTIFACTORY_URL` environment variable.
* `username` - (Optional) Username for basic auth. Requires `password` to be set.
    Conflicts with `api_key`, and `access_token`. This can also be sourced from the `ARTIFACTORY_USERNAME` environment variable.
* `password` - (Optional) Password for basic auth. Requires `username` to be set.
    Conflicts with `api_key`, and `access_token`. This can also be sourced from the `ARTIFACTORY_PASSWORD` environment variable.
* `api_key` - (Optional) API key for api auth. Uses `X-JFrog-Art-Api` header.
    Conflicts with `username`, `password`, and `access_token`. This can also be sourced from the `ARTIFACTORY_API_KEY` environment variable.
* `access_token` - (Optional) API key for token auth. Uses `Authorization: Bearer` header. For xray functionality, this is the only auth method accepted
    Conflicts with `username` and `password`, and `api_key`. This can also be sourced from the `ARTIFACTORY_ACCESS_TOKEN` environment variable.
* `check_license` - (Optional) Toggle for pre-flight checking of Artifactory license. Default to `true`.
