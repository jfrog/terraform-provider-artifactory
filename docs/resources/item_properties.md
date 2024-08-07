---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artifactory_item_properties Resource - terraform-provider-artifactory"
subcategory: "Artifact"
description: |-
  Provides a resource for managaing item (file, folder, or repository) properties. When a folder is used property attachment is recursive by default. See JFrog documentation https://jfrog.com/help/r/jfrog-artifactory-documentation/working-with-jfrog-properties for more details.
---

# artifactory_item_properties (Resource)

Provides a resource for managaing item (file, folder, or repository) properties. When a folder is used property attachment is recursive by default. See [JFrog documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/working-with-jfrog-properties) for more details.

## Example Usage

```terraform
resource "artifactory_item_properties" "my-repo-properties" {
  repo_key = "my-generic-local"
  properties = {
    "key1": ["value1"],
    "key2": ["value2", "value3"]
  }
  is_recursive = true
}

resource "artifactory_item_properties" "my-folder-properties" {
  repo_key = "my-generic-local"
  item_path = "folder/subfolder"
  properties = {
    "key1": ["value1"],
    "key2": ["value2", "value3"]
  }
  is_recursive = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `properties` (Map of Set of String) Map of key and list of values.

~>Keys are limited up to 255 characters and values are limited up to 2,400 characters. Using properties with values over this limit might cause backend issues.

~>The following special characters are forbidden in the key field: `)(}{][*+^$/~``!@#%&<>;=,±§` and the space character.
- `repo_key` (String) Respository key.

### Optional

- `is_recursive` (Boolean) Add this property to the selected folder and to all of artifacts and folders under this folder. Default to `false`
- `item_path` (String) The relative path of the item (file/folder/repository). Leave unset for repository.

## Import

Import is supported using the following syntax:

```shell
terraform import artifactory_item_properties.my-repo-properties repo_key

terraform import artifactory_item_properties.my-folder-properties repo_key:folder/subfolder
```
