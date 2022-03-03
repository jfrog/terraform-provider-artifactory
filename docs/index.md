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
  access_token = "${var.artifactory_access_token}"
}

# Create a new repository
resource "artifactory_local_repository" "pypi-libs" {
  key             = "pypi-libs"
  package_type    = "pypi"
  repo_layout_ref = "simple-default"
  description     = "A pypi repository for python packages"
}
```

## Authentication

The Artifactory provider supports Bearer Token authentication. 

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

## Argument Reference

The following arguments are supported:

* `url` - (Optional) URL of Artifactory. This can also be sourced from the `ARTIFACTORY_URL` environment variable.
* `access_token` - (Optional) API key for token auth. Uses `Authorization: Bearer` header. 
  This can also be sourced from the `ARTIFACTORY_ACCESS_TOKEN` environment variable.
* `check_license` - (Optional) Toggle for pre-flight checking of Artifactory license. Default to `true`.
