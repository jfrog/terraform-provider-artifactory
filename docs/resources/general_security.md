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
  enable_anonymous_access = true
}
```

## Argument Reference

The following arguments are supported:

* `enable_anonoymous_access` - (Optional) Enable anonymous access.  Default value is `false`.

## Import

Current general security settings can be imported using `security` as the `ID`, e.g.

```
$ terraform import artifactory_general_security.security security
```

~>The `artifactory_general_security` resource uses endpoints that are undocumented and may not work with SaaS
environments, or may change without notice.
