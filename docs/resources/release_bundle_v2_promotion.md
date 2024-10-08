---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artifactory_release_bundle_v2_promotion Resource - terraform-provider-artifactory"
subcategory: "Lifecycle"
description: |-
  This resource enables you to promote Release Bundle V2 version. For more information, see JFrog documentation https://jfrog.com/help/r/jfrog-artifactory-documentation/promote-a-release-bundle-v2-to-a-target-environment.
---

# artifactory_release_bundle_v2_promotion (Resource)

This resource enables you to promote Release Bundle V2 version. For more information, see [JFrog documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/promote-a-release-bundle-v2-to-a-target-environment).

## Example Usage

```terraform
resource "artifactory_release_bundle_v2_promotion" "my-release-bundle-v2-promotion" {
  name = "my-release-bundle-v2-artifacts"
  version = "1.0.0"
  keypair_name = "my-keypair-name"
  environment = "DEV"
  included_repository_keys = ["commons-qa-maven-local"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment` (String) Target environment
- `keypair_name` (String) Key-pair name to use for signature creation
- `name` (String) Name of Release Bundle
- `version` (String) Version to promote

### Optional

- `excluded_repository_keys` (Set of String) Defines specific repositories to exclude from the promotion.
- `included_repository_keys` (Set of String) Defines specific repositories to include in the promotion. If this property is left undefined, all repositories (except those specifically excluded) are included in the promotion. Important: If one or more repositories are specifically included, all other repositories are excluded (regardless of what is defined in `excluded_repository_keys`).
- `project_key` (String) Project key the Release Bundle belongs to

### Read-Only

- `created` (String) Timestamp when the new version was created (ISO 8601 standard).
- `created_millis` (Number) Timestamp when the new version was created (in milliseconds).
