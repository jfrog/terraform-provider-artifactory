---
subcategory: "Configuration"
---
# Artifactory OAuth SSO Settings Resource

This resource can be used to manage Artifactory's OAuth SSO settings.

Only a single `artifactory_oauth_settings` resource is meant to be defined.

~>The `artifactory_oauth_settings` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory OAuth SSO settings
resource "artifactory_oauth_settings" "oauth" {
  enable                        = true
  persist_users 	            = true
  allow_user_to_access_profile  = true

  oauth_provider {
    name 	      = "okta"
    enabled       = false
    type 	      = "openId"
    client_id     = "foo"
    client_secret = "bar"
    api_url       = "https://organization.okta.com/oauth2/v1/userinfo"
    auth_url      = "https://organization.okta.com/oauth2/v1/authorize"
    token_url     = "https://organization.okta.com/oauth2/v1/token"
  }
}
```

## Argument Reference

The following arguments are supported:

* `enable`                          - (Optional) Enable OAuth SSO.  Default value is `true`.
* `persist_users`                   - (Optional) Enable the creation of local Artifactory users.  Default value is `false`.
* `allow_user_to_access_profile`    - (Optional) Allow persisted users to access their profile.  Default value is `false`.
* `oauth_provider`                  - (Required) OAuth provider settings block. Multiple blocks can be defined, at least one is required.
    * `enabled`                     - (Optional) Enable the Artifactory OAuth provider.  Default value is `true`.
    * `name`                        - (Required) Name of the Artifactory OAuth provider.
    * `type`                        - (Required) Type of OAuth provider. (e.g., `github`, `google`, `cloudfoundry`, or `openId`)
    * `client_id`                   - (Required) OAuth client ID configured on the IdP.
    * `client_secret`               - (Required) OAuth client secret configured on the IdP.
    * `api_url`                     - (Required) OAuth user info endpoint for the IdP.
    * `auth_url`                    - (Required) OAuth authorization endpoint for the IdP.
    * `token_url`                   - (Required) OAuth token endpoint for the IdP.

## Import

Current OAuth SSO settings can be imported using `oauth_settings` as the `ID`.
If the resource is being imported, there will be a state drift, because `client_secret` can't be known. There are two options on how to approach this: 
1) Don't set `client_secret` initially, import, then update the config with actual secret;
2) Accept that there is a drift initially and run `terraform apply` twice;

```
$ terraform import artifactory_oauth_settings.oauth oauth_settings
```
