<a href="https://jfrog.com">
    <img src=".github/jfrog-logo-2022.svg" alt="JFrog logo" title="JFrog" align="right" height="50" />
</a>

# Terraform Provider Artifactory

[![Terraform & OpenTofu Acceptance Tests](https://github.com/jfrog/terraform-provider-artifactory/actions/workflows/acceptance-tests.yml/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions/workflows/acceptance-tests.yml)
[![Release Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

## Releases

Current provider major release: **12.x**

See [CHANGELOG.md](CHANGELOG.md) for full details

## Versions

Version 6.x is compatible with the Artifactory versions 7.49.x and below.

Version 7.x and 8.x is only compatible with Artifactory between 7.50.x and 7.67.x due to changes in the projects functionality.

Version 10.x (and later) is compatible with latest Artifactory versions (>=7.68.7 (self-hosted) and >=7.67.0 (cloud)).

## Terraform CLI version support

Current version support [Terraform Protocol v6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) which mean Terraform CLI version 1.0 and later. 

## Quick Start

Create a new Terraform file with `artifactory` resources. Also see [sample.tf](./sample.tf):

### HCL Example

```terraform
# Required for Terraform 1.0 and up (https://www.terraform.io/upgrade-guides)
terraform {
  required_providers {
    artifactory = {
      source  = "jfrog/artifactory"
      version = "12.3.3"
    }
  }
}

provider "artifactory" {
  // supply JFROG_ACCESS_TOKEN, and JFROG_URL as env vars
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

Initialize Terrform:
```console
terraform init
```

Plan (or Apply):
```console
terraform plan
```

## Documentation

To use this provider in your Terraform module, follow the documentation on [Terraform Registry](https://registry.terraform.io/providers/jfrog/artifactory/latest/docs).

## Versioning

In general, this project follows [Terraform Versioning Specification](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#versioning-specification) as closely as we can for tagging releases of the package.

## Developers Wiki

You can find building, testing and debugging information in the [Developers Wiki](https://github.com/jfrog/terraform-provider-artifactory/wiki) on GitHub.

## Contributors
See the [contribution guide](CONTRIBUTIONS.md).

## License

Copyright (c) 2025 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
