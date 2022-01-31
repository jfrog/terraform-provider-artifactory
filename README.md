# Terraform Provider Artifactory

[![Actions Status](https://github.com/jfrog/terraform-provider-artifactory/workflows/release/badge.svg)](https://github.com/jfrog/terraform-provider-artifactory/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/terraform-provider-artifactory)](https://goreportcard.com/report/github.com/jfrog/terraform-provider-artifactory)

To use this provider in your Terraform module, follow the documentation [here](https://registry.terraform.io/providers/jfrog/artifactory/latest/docs).

## License requirements:

This provider requires access to Artifactory APIs, which are only available in the _licensed_ pro and enterprise editions. You can determine which license you have by accessing the following URL `${host}/artifactory/api/system/licenses/`

You can either access it via api, or web browser - it does require admin level credentials, but it's one of the few APIs that will work without a license (side node: you can also install your license here with a `POST`)
```sh
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

## Limitations of functionality
The current way a repository is created is essentially through a union of fields of certain repo types.
It's important to note that, the official documentation is used only for inspiration as the documentation is quite wrong.
Support for some features has been achieved entirely through reverse engineering.

### Local repository limitations
[Local repository creation](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-LocalRepository) does not support (directly), repository specific fields in all cases. It's basically a union of
- Base repo params
- Maven
- Gradle
- Debian
- Docker (v1)
- RPM

### Remote repository limitations
[Remote repository creation](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-RemoteRepository) does not support (directly), repository specific fields in all cases. It's basically a union of
- base remote repo params
- bower support
- maven
- gradle
- Docker (v1)
- VCS
- Pypi
- Nuget

Query params may be forwarded, but this field doesn't exist in the documentation

### Permission target limitations
[Permission target V2](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplacePermissionTarget)

Permission target V1 support has been totally removed. Dynamically testing of permission targets using a new repository currently doesn't work because of race conditions when creating a repo. This will have to be resolved with retries at a later date.

### Changes to user creation
Previously, passwords were being generated for the user if none was supplied. This was both unnecessary (since TF has a password provider) and because the internal implementation could never be entirely in line with the remote server (or, be sure it was).

Now this functionality is removed. Password is a required field. The verification is offloaded to the Artifactory, which makes more sense, so we don't need to catch up with any possible changes on the Artifactory side.

## Build the Provider
Simply run `make install` - this will compile the provider and install it to `~/.terraform.d`. When running this, it will take the current tag and bump it 1 minor version. It does not actually create a new tag (that is `make release`). If you wish to use the locally installed provider, make sure your TF script refers to the new version number

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) 0.13
- [Go](https://golang.org/doc/install) 1.15+ (to build the provider plugin)

## Testing
How to run the tests isn't obvious.

First, you need a running instance of the JFrog platform (RT and XR). However, there is no currently supported dockerized, local version. You can ask for an instance to test against in as part of your PR. Alternatively, you can run the file [scripts/run-artifactory.sh](scripts/run-artifactory.sh), which, if have a file in the same directory called `artifactory.lic`, you can start just an artifactory instance. The license is not supplied, but a [30 day trial license can be freely obtained](https://jfrog.com/start-free/#hosted) and will allow local development.

Once you have that done you must set the following properties

Then, you have to set some environment variables as this is how the acceptance tests pick up their config
```sh
ARTIFACTORY_URL=http://localhost:8082
ARTIFACTORY_USERNAME=admin
ARTIFACTORY_PASSWORD=password
TF_ACC=true
```
A crucial, and very much hidden, env var to set is `TF_ACC=true` - you can literally set `TF_ACC` to anything you want, so long as it's set. The acceptance tests use terraform testing libraries that, if this flag isn't set, will skip all tests.

You can then run the tests as

```sh
$ go test -v ./pkg/...
```

Or

```sh
$ make acceptance
```

**DO NOT** remove the `-v` - terraform testing needs this (don't ask me why). This will recursively run all tests, including acceptance tests.

### Testing Federated repos

To execute acceptance tests for federated repo resource, we need:
- 2 Artifactory instances, configured with [Circle-of-Trust](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-CircleofTrust(Cross-InstanceAuthentication))
- Set environment variables `ARTIFACTORY_URL_2` to enable the acceptance tests that utilize both Artifactory instances

#### Setup Artifactory instances

The [scripts/run-artifactory.sh](scripts/run-artifactory.sh) starts two Artifactory instances for testing using the file [scripts/docker-compose.yml](scripts/docker-compose.yml).

`artifactory-1` is on the usual 8080/8081/8082 ports while `artifactory-2` is on 9080/9081/9082

#### Enable acceptance tests

Set the env var to the second Artifactory instance URL. This is the URL that will be accessible from `artifactory-1` container (not the URL from the Docker host):
```sh
$ export ARTIFACTORY_URL_2=http://artifactory-2:8081
```

Run all the acceptance tests as usual
```sh
$ make acceptance
```

Or run only the acceptance tests for Federated repos:
```sh
$ make acceptance_federated
```

## Debugging
Debugging a terraform provider is not straightforward. Terraform forks your provider as a separate process and then connects to it via RPC. Normally, when debugging, you would start the process to debug directly. However, with the terraform + go architecture, this isn't possible. So, you need to run terraform as you normally would and attach to the provider process by getting it's pid. This would be really tricky considering how fast the process can come up and be down. So, you need to actually halt the provider and have it wait for your debugger to attach.

Having said all that, here are the steps:
1. Install [delve](https://github.com/go-delve/delve)
2. Keep in mind that terraform will parallel process if it can, and it will start new instances of the TF provider process when running apply between the plan and confirmation.
   Add a snippet of go code to the close to where you need to break where in you install a busy sleep loop:
```go
	debug := true
	for debug {
		time.Sleep(time.Second) // set breakpoint here
	}
```
Then set a breakpoint inside the loop. Once you have attached to the process you can set the `debug` value to `false`, thus breaking the sleep loop and allow you to continue.
2. Compile the provider with debug symbology (`go build -gcflags "all=-N -l"`)
3. Install the provider (change as needed for your version)
```sh
# this will bump your version by 1 so it doesn't download from TF. Make sure you update any test scripts accordingly
make install
```
4. Run your provider: `terraform init && terraform plan` - it will start in this busy sleep loop.
5. In a separate shell, find the `PID` of the provider that got forked
`pgrep terraform-provider-artifactory`
6. Then, attach the debugger to that pid: `dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $pid`
A 1-liner for this whole process is:
`dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $(pgrep terraform-provider-artifactory)`
7. In intellij, setup a remote go debugging session (the default port is `2345`, but make sure it's set.) And click the `debug` button
8. Your editor should immediately break at the breakpoint from step 2. At this point, in the watch window, edit the `debug` value and set it to false, and allow the debugger to continue. Be ready for your debugging as this will release the provider and continue executing normally.

You will need to repeat steps 4-8 everytime you want to debug

## Versioning
In general, this project follows [semver](https://semver.org/) as closely as we can for tagging releases of the package. We've adopted the following versioning policy:

* We increment the **major version** with any incompatible change to functionality, including changes to the exported Go API surface or behavior of the API.
* We increment the **minor version** with any backwards-compatible changes to functionality.
* We increment the **patch version** with any backwards-compatible bug fixes.

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
