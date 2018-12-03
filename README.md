# Terraform Provider Artifactory #
## Using the provider ##

If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

## Requirements ##
-	[Go](https://golang.org/doc/install) 1.10+ (to build the provider plugin)
-	[Terraform](https://www.terraform.io/downloads.html) 0.11

## Building The Provider ##

Clone repository to: `$GOPATH/src/github.com/atlassian/terraform-provider-artifactory`

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/atlassian/terraform-provider-artifactory
go build
```

To install the provider
```sh
cd $GOPATH/src/github.com/atlassian/terraform-provider-artifactory
go install
```

## Roadmap ##

This library is being initially developed for an internal application at
Atlassian, so resources will likely be implemented in the order that they are
needed. Eventually, it would be ideal to cover the entire Artifactory API, so 
contributions are of course always welcome. The calling pattern is pretty well 
established, so adding new methods is relatively straightforward.

## Versioning ##

In general, this project follows [semver](https://semver.org/) as closely as we
can for tagging releases of the package. We've adopted the following versioning policy:

* We increment the **major version** with any incompatible change to
	functionality, including changes to the exported Go API surface
	or behavior of the API.
* We increment the **minor version** with any backwards-compatible changes to
	functionality.
* We increment the **patch version** with any backwards-compatible bug fixes.

## Reporting issues ##

We believe in open contributions and the power of a strong development community. Please read our [Contributing guidelines][CONTRIBUTING] on how to contribute back and report issues to terraform-provider-artifactory.

## Contributors ##

Pull requests, issues and comments are welcomed. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests
* Read [Contributing guidelines][CONTRIBUTING] for more details

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating
an issue and explaining the intended change.

Atlassian requires contributors to sign a Contributor License Agreement,
known as a CLA. This serves as a record stating that the contributor is
entitled to contribute the code/documentation/translation to the project
and is willing to have it used in distributions and derivative works
(or is willing to transfer ownership).

Prior to accepting your contributions we ask that you please follow the appropriate
link below to digitally sign the CLA. The Corporate CLA is for those who are
contributing as a member of an organization and the individual CLA is for
those contributing as an individual.

* [CLA for corporate contributors](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b)
* [CLA for individuals](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d)


## License ##
Copyright (c) 2017 Atlassian and others. Apache 2.0 licensed, see [LICENSE][LICENSE] file.


[CONTRIBUTING]: .github/CONTRIBUTING.md
[LICENSE]: ./LICENSE.txt