# Contribution Guide

## Contributors

Pull requests, issues and comments are welcomed. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating an issue and explaining the intended change.

JFrog requires contributors to sign a Contributor License Agreement, known as a CLA. This serves as a record stating that the contributor is entitled to contribute the code/documentation/translation to the project and is willing to have it used in distributions and derivative works (or is willing to transfer ownership).

## Building

Simply run `make install` - this will compile the provider and install it to `~/.terraform.d`. When running this, it will take the current tag and bump it 1 patch version. It does not actually create a new tag (that will be `make release`). If you wish to use the locally installed provider, make sure your TF script refers to the new version number.

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- [Go](https://golang.org/doc/install) 1.18+ (to build the provider plugin)

## Debugging

See [debugging wiki](https://github.com/jfrog/terraform-provider-artifactory/wiki/Debugging).

## Testing

First, you need a running instance of the JFrog Artifactory.

You can ask for an instance to test against it as part of your PR. Alternatively, you can run the file [scripts/run-artifactory.sh](scripts/run-artifactory.sh).

The script requires a valid license of a [supported type](https://github.com/jfrog/terraform-provider-artifactory#license-requirements), license should be saved in the file called `artifactory.lic` in the same directory as a script.

With the script you can start one or two Artifactory instances using docker compose.

The license is not supplied, but a [30 day trial license can be freely obtained](https://jfrog.com/start-free/#hosted) and will allow local development. Make sure the license saved as a multi line text file.

Currently, acceptance tests **require an access key** and don't support basic authentication or an API key. To generate an access key, please refer to the [official documentation](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-GeneratingAdminTokens)

Then, you have to set some environment variables as this is how the acceptance tests pick up their config.

```sh
ARTIFACTORY_URL=http://localhost:8082
ARTIFACTORY_USERNAME=admin
ARTIFACTORY_ACCESS_TOKEN=<your_access_token>
TF_ACC=true
```
`ARTIFACTORY_USERNAME` is not used in authentication, but used in several tests, related to replication functionality. It should be hardcoded to `admin`, because it's a default user created in the Artifactory instance from the start.

A crucial env var to set is `TF_ACC=true` - you can literally set `TF_ACC` to anything you want, so long as it's set. The acceptance tests use terraform testing libraries that, if this flag isn't set, will skip all tests.

You can then run the tests as

```sh
$ go test -v -p 1 ./pkg/...
```

Or

```sh
$ make acceptance
```

**DO NOT** remove the `-v` - terraform testing needs this. This will recursively run all tests, including acceptance tests.

### Testing Federated repos

To execute acceptance tests for federated repo resource, we need:
- 2 Artifactory instances, configured with [Circle-of-Trust](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-CircleofTrust(Cross-InstanceAuthentication))
- Set environment variables `ARTIFACTORY_URL_2` to enable the acceptance tests that utilize both Artifactory instances

#### Setup Artifactory instances

The [scripts/run-artifactory.sh](scripts/run-artifactory.sh) starts two Artifactory instances for testing using the file [scripts/docker-compose.yml](scripts/docker-compose.yml).

`artifactory-1` is on the usual 8080/8081/8082 ports while `artifactory-2` is on 9080/9081/9082

You can also run one instance of artifactory using `./run-artifactory-container.sh` which doesn't use docker-compose.

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

## Releasing

Please create a pull request against the master branch. Each pull request will be reviewed by a member of the JFrog team.

#### Thank you for contributing!
