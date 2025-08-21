---
subcategory: "Configuration"
---
# Artifactory General Security Resource

This resource can be used to manage Artifactory's general security settings.

Only a single `artifactory_general_security` resource is meant to be defined.

~>The `artifactory_general_security` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory general security settings
resource "artifactory_general_security" "security" {
  enable_anonymous_access = false
  encryption_policy       = "REQUIRED"
}
```

```hcl
# Enable anonymous access with supported encryption
resource "artifactory_general_security" "security" {
  enable_anonymous_access = true
  encryption_policy       = "SUPPORTED"
}
```

```hcl
# Disable encryption policy (clear-text only)
resource "artifactory_general_security" "security" {
  enable_anonymous_access = false
  encryption_policy       = "UNSUPPORTED"
}
```

## Argument Reference

The following arguments are supported:

* `enable_anonymous_access` - (Optional) Enable anonymous access. Default value is `false`.
* `encryption_policy` - (Optional) Determines the password requirements from users identified to Artifactory from a remote client such as Maven. The options are:
  - `SUPPORTED` (default): Users can authenticate using secure encrypted passwords or clear-text passwords.
  - `REQUIRED`: Users must authenticate using secure encrypted passwords. Clear-text authentication fails.
  - `UNSUPPORTED`: Only clear-text passwords can be used for authentication.
  
  Default value is `SUPPORTED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the general security configuration (always `security`).

## Import

Current general security settings can be imported using `security` as the `ID`, e.g.

```
$ terraform import artifactory_general_security.security security
```

~>The `artifactory_general_security` resource uses endpoints that are undocumented and may not work with SaaS environments, or may change without notice.
