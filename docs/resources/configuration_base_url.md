---
subcategory: "Configuration"
---
# Artifactory Custom Base URL Resource

This resource can be used to update/change Artifactory's Custom Base URL.

Only a single `artifactory_configuration_base_url` resource is meant to be defined.

~>The `artifactory_configuration_base_url` resource manages the Base URL configuration in Artifactory.

## Example Usage

```hcl
# Configure Artifactory Custom Base Url
resource "artifactory_configuration_base_url" "baseurl" {
  base_url = "http://localhost:8082"
}
```

## Argument Reference

The following argument are supported:

* `baseurl`         - (required) The Base URL for Artifactory.
