# Terraform Provider Artifactory

[![Actions Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

To use this provider in your Terraform module, follow the documentation [here](https://registry.terraform.io/providers/jfrog/artifactory/latest/docs).

## Important note:

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
## Testing
How to run the tests isn't obvious.
First, you need a running instance of artifactory. Included with this repo is a functioning `docker-compose.yml`
that will fire up the whole stack for you. Run it as:

```bash
docker-compose --env-file .env up -d
```
Then, you have to set some environment variables as this is how the acceptance tests pick up their config
```bash
ARTIFACTORY_URL=http://localhost:8082
ARTIFACTORY_USERNAME=admin
ARTIFACTORY_PASSWORD=password
TF_ACC=FOO
```
a crucial, and very much hidden, env var to set is
`TF_ACC=FOO` - you can literally set `TF_ACC` to anything you want, so long as it's set. The acceptance tests use
terraform testing libraries that, if this flag isn't set, will skip all tests.

You can then run the tests as
`go test -v ./pkg/...`
**DO NOT** remove the `-v` - terraform testing needs this (don't ask me why). This will recursively run all tests, including
acceptance tests. 

## Debugging
Debugging a terraform provider is not straightforward. Terraform forks your provider as a separate process and then 
connects to it via RPC. Normally, when debugging, you would start the process to debug directly. However, with the 
terraform + go architecture, this isn't possible. So, you need to run terraform as you normally would and attach to the
provider process by getting it's pid. This would be really tricky considering how fast the process can come up and be down.
So, you need to actually halt the provider and have it wait for your debugger to attach. 

Having said all that, here are the steps:
1. Install [delve](https://github.com/go-delve/delve)
2. Add a snippet of go code to the [provider initializer](pkg/artifactory/provider.go) `providerConfigure`, where in you install a busy sleep loop:
```go
	debug := true
	for debug {
		time.Sleep(time.Second) // set breakpoint here
	}
``` 
and set a breakpoint inside the loop. Once you have attached to the process you can set the `debug` value to `false`,
thus breaking the sleep loop and allow you to continue. 
2. Compile the provider with debug symbology (`go build -gcflags "all=-N -l"`)
3. Install the provider (change as needed for your version)
```bash 
mkdir -p .terraform/plugins/registry.terraform.io/jfrog/artifactory/2.2.5/darwin_amd64 \
    && mv terraform-provider-artifactory .terraform/plugins/registry.terraform.io/jfrog/artifactory/2.2.5/darwin_amd64
```
4. Run your provider: `terraform init && terraform plan` - it will start in this busy sleep loop.
5. In a separate shell, find the `PID` of the provider that got forked 
`pgrep terraform-provider-artifactory`
6. Then, attach the debugger to that pid: `dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $pid`
A 1-liner for this whole process is: 
`dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $(pgrep terraform-provider-artifactory)`
7. In intellij, setup a remote go debugging session (the default port is `2345`, but make sure it's set.) And click the `debug` button
8. Your editor should immediately break at the breakpoint from step 2. At this point, in the watch window, edit the `debug` 
value and set it to false, and allow the debugger to continue. Be ready for your debugging as this will release the provider 
and continue executing normally.

You will need to repeat steps 4-8 everytime you want to debug





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
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/jfrog/terraform)