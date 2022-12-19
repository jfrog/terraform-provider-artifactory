---
subcategory: "Configuration"
---
# Artifactory SAML SSO Settings Resource

This resource can be used to manage Artifactory's SAML SSO settings.

Only a single `artifactory_saml_settings` resource is meant to be defined.

~>The `artifactory_saml_settings` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory SAML SSO settings
resource "artifactory_saml_settings" "saml" {
  enable                       = true
  service_provider_name        = "okta"
  login_url                    = "test-login-url"
  logout_url                   = "test-logout-url"
  certificate                  = "test-certificate"
  email_attribute              = "email"
  group_attribute              = "groups"
  no_auto_user_creation        = false
  allow_user_to_access_profile = true
  auto_redirect                = true
  sync_groups                  = true
  verify_audience_restriction  = true
  use_encrypted_assertion      = false
}
```

## Argument Reference

The following arguments are supported:

* `enable`                          - (Optional) Enable SAML SSO.  Default value is `true`.
* `service_provider_name`           - (Required) The SAML service provider name. This should be a URI that is also known as the entityID, providerID, or entity identity.
* `login_url`                       - (Required) Service provider login url configured on the IdP.
* `logout_url`                      - (Required) Service provider logout url, or where to redirect after user logs out.
* `certificate`                     - (Optional) SAML certificate that contains the public key for the IdP service provider.  Used by Artifactory to verify sign-in requests. Default value is ``.
* `email_attribute`                 - (Optional) Name of the attribute in the SAML response from the IdP that contains the user's email. Default value is ``.
* `group_attribute`                 - (Optional) Name of the attribute in the SAML response from the IdP that contains the user's group memberships. Default value is ``.  
* `no_auto_user_creation`           - (Optional) When automatic user creation is off, authenticated users are not automatically created inside Artifactory. Instead, for every request from an SSO user, the user is temporarily associated with default groups (if such groups are defined), and the permissions for these groups apply. Without auto-user creation, you must manually create the user inside Artifactory to manage user permissions not attached to their default groups. Default value is `false`.
* `allow_user_to_access_profile`    - (Optional) Allow persisted users to access their profile.  Default value is `true`.
* `auto_redirect`                   - (Optional) Auto redirect to login through the IdP when clicking on Artifactory's login link.  Default value is `false`.
* `sync_groups`                     - (Optional) Associate user with Artifactory groups based on the `group_attribute` provided in the SAML response from the identity provider.  Default value is `false`.
* `verify_audience_restriction`     - (Optional) Enable "audience", or who the SAML assertion is intended for.  Ensures that the correct service provider intended for Artifactory is used on the IdP.  Default value is `true`.
* `use_encrypted_assertion`         - (Optional) When set, an X.509 public certificate will be created by Artifactory. Download this certificate and upload it to your IDP and choose your own encryption algorithm. This process will let you encrypt the assertion section in your SAML response. Default value is `false`.

## Import

Current SAML SSO settings can be imported using `saml_settings` as the `ID`, e.g.

```
$ terraform import artifactory_saml_settings.saml saml_settings
```
