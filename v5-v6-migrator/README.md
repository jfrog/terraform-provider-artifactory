# Artifactory Terraform provider v5 to v6 repository resource migrator

A CLI tool to transform Terraform resource for Artifactory repository from generic type in V5 to package specific type in V6.

The tool reads and parse the input Terraform configuration file and create a new file with the V5 resources replaced by V6 resources. Additionally this tool can output Terraform state import commands for the migrated resources to help the migration process.

## Usage

```sh
tf-v5-migrator --input sample.v5.tf --output sample.v6.tf
```

To include Terraform import statements in the output, use the `--import` flag

```sh
tf-v5-migrator --input sample.v5.tf --output sample.v6.tf --import
```

Will output:
```sh
terraform import artifactory_local_npm_repository.alexh-npm-local-2 alexh-npm-local-2-key
terraform import artifactory_remote_npm_repository.alexh-npm-remote alexh-npm-remote-key
terraform import artifactory_remote_npm_repository.alexh-npm-remote-2 alexh-npm-remote-2-key
terraform import artifactory_virtual_npm_repository.alexh-npm-virtual alexh-npm-virtual-key
terraform import artifactory_virtual_npm_repository.alexh-npm-virtual-2 alexh-npm-virtual-2-key
```

## Build

### Pre-requisites

* Go 1.18

To build the binary, run build command in shell:

```sh
make build
```

This will create a binary in the `./bin` directory.

## Contributors
See the [contribution guide](../CONTRIBUTIONS.md).

## License

Copyright (c) 2023 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ../LICENSE
