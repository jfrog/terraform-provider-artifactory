<a href="https://jfrog.com">
    <img src=".github/jfrog-logo-2022.svg" alt="JFrog logo" title="JFrog" align="right" height="50" />
</a>

# Terraform Provider Artifactory

[![Actions Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

## Releases

Current major release: **6.x**

See [CHANGELOG.md](CHANGELOG.md) for full details

<details><summary>Recent Releases</summary>
### 6.6.0

IMPROVEMENTS:

* resource/artifactory_group: Add `external_id` attribute to support Azure AD group. PR: [#437](https://github.com/jfrog/terraform-provider-artifactory/pull/437). Issue [#429](https://github.com/jfrog/terraform-provider-artifactory/issues/429)

### 6.5.3

IMPROVEMENTS:

* reorganizing documentation, adding missing documentation links, fixing formatting. No changes in the functionality.
PR: [GH-435](https://github.com/jfrog/terraform-provider-artifactory/pull/435). Issues [#422](https://github.com/jfrog/terraform-provider-artifactory/issues/422) and [#398](https://github.com/jfrog/terraform-provider-artifactory/issues/398)

### 6.5.2

IMPROVEMENTS:

* resource/artifactory_artifact_webhook: Added 'cached' event type for Artifact webhook. PR: [GH-430](https://github.com/jfrog/terraform-provider-artifactory/pull/430).

### 6.5.1

BUG FIXES:

* provider:  Setting the right default value for 'access_token' attribute. PR: [GH-426](https://github.com/jfrog/terraform-provider-artifactory/pull/426). Issue [#425](https://github.com/jfrog/terraform-provider-artifactory/issues/425)

### 6.5.0

IMPROVEMENTS:

* Resources added for Pub package type of Local Repository
* Resources added for Pub package type of Remote Repository
* Resources added for Pub package type of Virtual Repository
* Acceptance test case enhanced with Client TLS Certificate

PR: [GH-421](https://github.com/jfrog/terraform-provider-artifactory/pull/421)
</details>

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

```js
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

## Contributors

Pull requests, issues and comments are welcomed. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating an issue and explaining the intended change.

JFrog requires contributors to sign a Contributor License Agreement, known as a CLA. This serves as a record stating that the contributor is entitled to contribute the code/documentation/translation to the project and is willing to have it used in distributions and derivative works (or is willing to transfer ownership).

## License

Copyright (c) 2022 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
