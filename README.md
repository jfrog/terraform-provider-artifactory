# Terraform Provider Artifactory
[![Actions Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/build/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

To use this provider in your Terraform module, follow the documentation [here](docs/index.md).

## Build the Provider
If you're building the provider, follow the instructions to [install it as a plugin](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin).
After placing it into your plugins directory,  run `terraform init` to initialize it.

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) 0.11
- [Go](https://golang.org/doc/install) 1.11+ (to build the provider plugin)

Clone repository to: `$GOPATH/src/github.com/jfrog/terraform-provider-artifactory`

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/jfrog/terraform-provider-artifactory
go build
```

To install the provider
```sh
cd $GOPATH/src/github.com/jfrog/terraform-provider-artifactory
go install
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

[Sign the CLA](https://cla-assistant.io/jfrog/terraform-provider-artifactory)

## License
Copyright (c) 2019 Atlassian and others.  
Copyright (c) 2020 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
