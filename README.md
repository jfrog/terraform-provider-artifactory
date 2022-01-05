# Terraform Provider Xray

[![Actions Status](https://github.com/jfrog/terraform-provider-xray/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-xray/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-xray)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-xray)

To use this provider in your Terraform module, follow the documentation [here](https://registry.terraform.io/providers/jfrog/xray/latest/docs).

[Xray general information](https://jfrog.com/xray/)

[Xray API Documentation](https://www.jfrog.com/confluence/display/JFROG/Xray+REST+API)

## Quick Start

Create a new Terraform file with `xray` resource (and `artifactory` resource as well):

<details><summary>HCL Example</summary>

```terraform
# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.24"
    }
    xray = {
      source  = "registry.terraform.io/jfrog/xray"
      version = "0.0.1"
    }
  }
}
provider "artifactory" {
  // supply ARTIFACTORY_USERNAME, ARTIFACTORY_PASSWORD and ARTIFACTORY_URL as env vars
}

provider "xray" {
// Also user can supply the following env vars:
// JFROG_URL or XRAY_URL
// XRAY_ACCESS_TOKEN or JFROG_ACCESS_TOKEN
}

resource "random_id" "randid" {
  byte_length = 2
}

resource "artifactory_user" "user1" {
  name     = "user1"
  email    = "test-user1@artifactory-terraform.com"
  groups   = ["readers"]
  password = "Passw0rd!"
}

resource "artifactory_local_docker_v2_repository" "docker-local" {
  key             = "docker-local"
  description     = "hello docker-local"
  tag_retention   = 3
  max_unique_tags = 5
  xray_index = true # must be set to true to be able to assign the watch to the repo
}

resource "artifactory_local_gradle_repository" "local-gradle-repo" {
  key                             = "local-gradle-repo-basic"
  checksum_policy_type            = "client-checksums"
  snapshot_version_behavior       = "unique"
  max_unique_snapshots            = 10
  handle_releases                 = true
  handle_snapshots                = true
  suppress_pom_consistency_checks = true
  xray_index = true # must be set to true to be able to assign the watch to the repo
}

resource "xray_security_policy" "security1" {
  name        = "test-security-policy-severity-${random_id.randid.dec}"
  description = "Security policy description"
  type        = "security"

  rule {
    name     = "rule-name-severity"
    priority = 1

    criteria {
      min_severity = "High"
    }

    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false // set to true only if Jira integration is enabled
      build_failure_grace_period_in_days = 5     // use only if fail_build is enabled

      block_download {
        unscanned = true
        active    = true
      }
    }
  }
}

resource "xray_security_policy" "security2" {
  name        = "test-security-policy-cvss-${random_id.randid.dec}"
  description = "Security policy description"
  type        = "security"

  rule {
    name     = "rule-name-cvss"
    priority = 1

    criteria {

      cvss_range {
        from = 1.5
        to   = 5.3
      }
    }

    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false // set to true only if Jira integration is enabled
      build_failure_grace_period_in_days = 5     // use only if fail_build is enabled

      block_download {
        unscanned = true
        active    = true
      }
    }
  }
}

resource "xray_license_policy" "license1" {
  name        = "test-license-policy-allowed-${random_id.randid.dec}"
  description = "License policy, allow certain licenses"
  type        = "license"

  rule {
    name     = "License_rule"
    priority = 1

    criteria {
      allowed_licenses         = ["Apache-1.0", "Apache-2.0"]
      allow_unknown            = false
      multi_license_permissive = true
    }

    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_release_bundle_distribution  = false
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false // set to true only if Jira integration is enabled
      custom_severity                    = "High"
      build_failure_grace_period_in_days = 5 // use only if fail_build is enabled

      block_download {
        unscanned = true
        active    = true
      }
    }
  }
}

resource "xray_license_policy" "license2" {
  name        = "test-license-policy-banned-${random_id.randid.dec}"
  description = "License policy, block certain licenses"
  type        = "license"

  rule {
    name     = "License_rule"
    priority = 1

    criteria {
      banned_licenses          = ["Apache-1.1", "APAFML"]
      allow_unknown            = false
      multi_license_permissive = false
    }

    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_release_bundle_distribution  = false
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false // set to true only if Jira integration is enabled
      custom_severity                    = "Medium"
      build_failure_grace_period_in_days = 5 // use only if fail_build is enabled

      block_download {
        unscanned = true
        active    = true
      }
    }
  }
}

resource "xray_watch" "all-repos" {
  name        = "all-repos-watch-${random_id.randid.dec}"
  description = "Watch for all repositories, matching the filter"
  active      = true

  watch_resource {
    type = "all-repos"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  assigned_policy {
    name = xray_security_policy.security1.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.license1.name
    type = "license"
  }
  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "repository" {
  name        = "repository-watch-${random_id.randid.dec}"
  description = "Watch a single repo or a list of repositories"
  active      = true

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = artifactory_local_docker_v2_repository.docker-local.key

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = artifactory_local_gradle_repository.local-gradle-repo.key

    filter {
      type  = "package-type"
      value = "Docker"
    }
  }

  assigned_policy {
    name = xray_security_policy.security1.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.license1.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "build" {
  name        = "build-watch-${random_id.randid.dec}"
  description = "Watch a single build or a list of builds"
  active      = true

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-build-name"
  }

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-other-build-name"
  }

  assigned_policy {
    name = xray_security_policy.security1.name
    type = "security"
  }
  assigned_policy {
    name = xray_license_policy.license1.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}
```


</details>


## Release notes for v0.0.1
Xray provider was separated from Artifactory provider. The most notable differences in the new Xray provider: 
- Provider uses Xray API v2 for all the API calls.
- HCL was changed and now uses singular names instead of the plurals for the repeatable elements, like `rule`, `watch_resource`, `filter` and `assigned_policy`.
- Security policy and License policy now are separate Terraform provider resources.
- In Schemas, TypeList was replaced by TypeSet (where it makes sense) to avoid sorting problems, when Terraform detect the change in sorted elements.
- Added multiple validations for Schemas to verify the data on the Terraform level instead of getting errors in the API response.


## License requirements:
This provider requires Xray to be added to your Artifactory installation. 
Xray requires minimum Pro Team license (Public Marketplace version or SaaS) or Pro X license (Self-hosted).
See the details [here](https://jfrog.com/pricing/#sass)
You can determine which license you have by accessing the following Artifactory URL `${host}/artifactory/api/system/licenses/`

## Limitations of functionality
Currently, Xray provider is not supporting JSON objects in the Watch filter value. We are working on adding this functionality. 


## Build the Provider
Simply run `make install` - this will compile the provider and install it to `~/.terraform.d`. When running this, it will
take the current tag and bump it 1 minor version. It does not actually create a new tag (that is `make release`).
If you wish to use the locally installed provider, make sure your TF script refers to the new version number.

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) 0.13
- [Go](https://golang.org/doc/install) 1.15+ (to build the provider plugin)

### Building on macOS

This provider uses [GNU sed](https://www.gnu.org/software/sed/) as part of the build toolchain, in both Linux and macOS. This provides consistency across OSes.

If you are building this on macOS, you have two options:
- Install [gnu-sed using brew](https://formulae.brew.sh/formula/gnu-sed), OR
- Use a Linux Docker image/container

#### Using gnu-sed

After installing with brew, get the GNU sed information:

```sh
$ brew info gnu-sed
```

You should see something like:
```
GNU "sed" has been installed as "gsed".
If you need to use it as "sed", you can add a "gnubin" directory
to your PATH from your bashrc like:

     PATH="$(brew --prefix)/opt/gnu-sed/libexec/gnubin:$PATH"
```

Add the `gnubin` directory to your `.bashrc` or `.zshrc` per instruction so that `sed` command uses gnu-sed.


## Testing
Since JFrog Xray is an addon for Artifactory, you will need a running instance of the JFrog platform (Artifactory and Xray).
However, there is no currently supported dockerized, local version. The fastest way to install Artifactory and Xray as a self-hosted installation is to use Platform
Helm chart. Free 30 days trial version is available [here](https://jfrog.com/start-free/#hosted) 
If you want to test on SaaS instance - [30 day trial can be freely obtained](https://jfrog.com/start-free/#saas) 
and will allow local development. 

Then, you have to set some environment variables as this is how the acceptance tests pick up their config:
```bash
JFROG_URL=http://localhost:8081
XRAY_ACCESS_TOKEN=your-admin-key
TF_ACC=true
```
a crucial, and very much hidden, env var to set is
`TF_ACC=true` - you can literally set `TF_ACC` to anything you want, so long as it's set. The acceptance tests use
terraform testing libraries that, if this flag isn't set, will skip all tests.

`XRAY_ACCESS_TOKEN` can be generated in the UI. Go to **Settings -> Identity and Access -> Access Tokens -> Generate Admin Token**


You can then run the tests as `make acceptance`. You can check what it's doing on the background in the [GNUmakefile](GNUmakefile) in the project. 

We've found that it's very convenient to use [Charles proxy](https://www.charlesproxy.com/) to see the payload, generated by Terraform Provider during the testing process.
You can also use any other network packet reader, like Wireshark and so on. 


## Registry documentation generation
All the documentation in the project is generated by [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).
If you make any changes to the resource schemas, you will need to re-generate documentation.
Install [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs#installation), then run:
```sh
$ make doc
```

## Versioning
In general, this project follows [semver](https://semver.org/) as closely as we
can for tagging releases of the package. We've adopted the following versioning policy:

* We increment the **major version** with any incompatible change to
  functionality, including changes to the exported Go API surface
  or behavior of the API.
* We increment the **minor version** with any backwards-compatible changes to
  functionality.
* We increment the **patch version** with any backwards-compatible bug fixes.

## Contributors
Pull requests, issues and comments are welcomed. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating
an issue and explaining the intended change.

JFrog requires contributors to sign a Contributor License Agreement,
known as a CLA. This serves as a record stating that the contributor is
entitled to contribute the code/documentation/translation to the project
and is willing to have it used in distributions and derivative works
(or is willing to transfer ownership).

## License
Copyright (c) 2021 JFrog.

Apache 2.0 licensed, see [LICENSE](LICENSE) file.