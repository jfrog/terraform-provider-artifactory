# Artifactory General Security Resource

This resource can be used to manage Artifactory's general security settings. Only a single `artifactory_general_security` resource is meant to be defined.

## Example Usage

```hcl
# Configure Artifactory general security settings
resource "artifactory_general_security" "security" {
  enable_anonymous_access = true
}
```

## Argument Reference

The following arguments are supported:

* `enable_anonymous_access` - (Optional) Enable anonymous READ access. The default value is `false`.

## Import

Current general security settings can be imported using `security` as the `ID`.

```
terraform import artifactory_general_security.security security
```
