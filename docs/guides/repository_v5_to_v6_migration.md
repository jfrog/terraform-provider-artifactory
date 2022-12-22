---
page_title: "Migrate repository resources from V5 to V6 provider"
---

# Migrate repository resources from V5 to V6 provider

The migration strategy we recommend consists of two stages:
1. In provider v5, migrate repository resources from 'generic' resources (`artifactory_local_repository`, `artifactory_remote_repository`, `artifactory_virtual_repository`) to package type specific resources (e.g. `artifactory_local_npm_repository`, `artifactory_remote_npm_repository`, `artifactory_virtual_npm_repository`).
2. Upgrade Artifactory provider to v6.x

## Migrate resources

~>Before performing any major changes to Terraform configuration, it's highly recommended to backup your Terraform state.

The major steps to migrate resources to new resource type are:
1. Define new resources using package type specific resources in your configuration to match your infrastructure. This can be as straightforward as copying the attributes from existing resources to new resources and removing the package_type attribute.
2. Import your repositories into Terraform state for these new resources
  ```sh
  $ terraform import artifactory_local_npm_repository.my-npm-local my-npm-local
  ```
3. Remove the original resource from Terraform state
  ```sh
  $ terraform state rm artifactory_local_repository.my-npm-local
  ```

## Upgrade provider to v6

While it may be tempting to jump straight to the latest version (v6.21.3), there have been numerous new features and bug fixes that may make this large jump challenging to tackle. For this reason, we suggest picking an earlier version of v6 as the initial upgrade target. This should enable a smoother process and allow you to gain trust and confidence before upgrading to newer versions. Use `CHANGELOG.md` to guide you in this process.

### Upgrade provider

Change provider version in `.tf` file

```hcl
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "6.21.3"
    }
  }
}
```

Initialize terraform working directory and upgrade the provider.

```sh
$ terraform init -upgrade
```
