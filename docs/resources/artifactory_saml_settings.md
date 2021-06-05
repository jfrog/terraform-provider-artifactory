# Artifactory SAML SSO Settings Resource

This resource can be used to manage Artifactory's SAML SSO settings.

Only a single `artifactory_saml_settings` resource is meant to be defined.

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
}
```

## Argument Reference

The following arguments are supported:

* `enable`                          - (Optional) Enable SAML SSO.  Default value is `true`.
* `service_provider_name`           - (Required) Name of the service provider configured on the .
* `login_url`                       - (Required) Service provider login url configured on the IdP.
* `logout_url`                      - (Required) Service provider logout url, or where to redirect after user logs out.
* `certificate`                     - (Optional) SAML certificate that contains the public key for the IdP service provider.  Used by Artifactory to verify sign-in requests.
* `email_attribute`                 - (Optional) Name of the attribute in the SAML response from the IdP that contains the user's email.
* `group_attribute`                 - (Optional) Name of the attribute in the SAML response from the IdP that contains the user's group memberships.  
* `no_auto_user_creation`           - (Optional) Enable the creation of local Artifactory users.
* `allow_user_to_access_profile`    - (Optional) Allow persisted users to access their profile.
* `auto_redirect`                   - (Optional) Auto redirect to login through the IdP when clicking on Artifactory's login link.
* `sync_groups`                     - (Optional) Associate user with Artifactory groups based on the `group_attribute` provided in the SAML response from the identity provider.
* `verify_audience_restriction`     - (Optional) Enable "audience", or who the SAML assertion is intended for.  Ensures that the correct service provider intended for Artifactory is used on the IdP.  

## Import

Current SAML SSO settings can be imported using `saml_settings` as the `ID`, e.g.

```
$ terraform import artifactory_saml_settings.saml saml_settings
```
