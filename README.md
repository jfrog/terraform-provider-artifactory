<a href="https://jfrog.com">
    <img src=".github/jfrog-logo-2022.svg" alt="JFrog logo" title="JFrog" align="right" height="50" />
</a>

# Terraform Provider Artifactory

[![Actions Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

## Releases

Current provider major release: **9.x**

See [CHANGELOG.md](CHANGELOG.md) for full details

## Versions

Version 6.x is compatible with the Artifactory versions 7.49.x and below.

Version 7.x and 8.x is only compatible with Artifactory between 7.50.x and 7.67.x due to changes in the projects functionality.

Version 9.x is the latest major version and is compatible with latest Artifactory versions (>=7.68.7 (self-hosted) and >=7.67.0 (cloud)).

## Terraform CLI version support

Current version support [Terraform Protocol v5](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-5) which mean Terraform CLI version 0.12 and later. 

~>We will be moving to [Terraform Protocol v6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) in **Q1 2024**. This means only Terraform CLI version 1.0 and later will be supported.

## Quick Start

Create a new Terraform file with `artifactory` resources. Also see [sample.tf](./sample.tf):

<details><summary>HCL Example</summary>

```terraform
# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "6.6.1"
    }
  }
}

provider "artifactory" {
  // supply ARTIFACTORY_USERNAME, ARTIFACTORY_ACCESS_TOKEN, and ARTIFACTORY_URL as env vars
}

resource "artifactory_local_pypi_repository" "pypi-local" {
  key         = "pypi-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_artifact_webhook" "artifact-webhook" {
  key         = "artifact-webhook"
  event_types = ["deployed", "deleted", "moved", "copied"]
  criteria {
    any_local        = true
    any_remote       = false
    repo_keys        = [artifactory_local_pypi_repository.pypi-local.key]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url    = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy  = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }

  depends_on = [artifactory_local_pypi_repository.pypi-local]
}
```
</details>

Initialize Terrform:
```sh
$ terraform init
```

Plan (or Apply):
```sh
$ terraform plan
```

## Documentation

To use this provider in your Terraform module, follow the documentation on [Terraform Registry](https://registry.terraform.io/providers/jfrog/artifactory/latest/docs).

## License requirements

This provider requires access to Artifactory APIs, which are only available in the _licensed_ pro and enterprise editions. You can determine which license you have by accessing the following URL `${host}/artifactory/api/system/licenses/`

You can either access it via API, or web browser - it requires admin level credentials, but it's one of the few APIs that will work without a license (side node: you can also install your license here with a `POST`)

```sh
$ curl -sL ${host}/artifactory/api/system/licenses/ | jq .
```

```json
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

## Versioning

In general, this project follows [Terraform Versioning Specification](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#versioning-specification) as closely as we can for tagging releases of the package.

## Developers Wiki

You can find building, testing and debugging information in the [Developers Wiki](https://github.com/jfrog/terraform-provider-artifactory/wiki) on GitHub.

## Contributors
See the [contribution guide](CONTRIBUTIONS.md).

## License

Copyright (c) 2023 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
