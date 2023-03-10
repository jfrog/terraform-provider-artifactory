---
subcategory: "Security"
---
# Artifactory Scoped Token Resource

Provides an Artifactory Scoped Token resource. This can be used to create and manage Artifactory Scoped Tokens.

!>Scoped Tokens will be stored in the raw state as plain-text. [Read more about sensitive data in
state](https://www.terraform.io/docs/state/sensitive-data.html).

~>Token would not be saved by Artifactory if `expires_in` is less than the persistency threshold value (default to 10800 seconds) set in Access configuration. See [Persistency Threshold](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-PersistencyThreshold) for details.

## Example Usages
### Create a new Artifactory scoped token for an existing user

```hcl
resource "artifactory_scoped_token" "scoped_token" {
  username = "existing-user"
}
```

**Note:** This assumes that the user `existing-user` has already been created in Artifactory by different means, i.e. manually or in a separate terraform apply.

### Create a new Artifactory user and scoped token
```hcl
resource "artifactory_user" "new_user" {
  name   = "new_user"
  email  = "new_user@somewhere.com"
  groups = ["readers"]
}

resource "artifactory_scoped_token" "scoped_token_user" {
  username = artifactory_user.new_user.name
}
```

### Creates a new token for groups
```hcl
resource "artifactory_scoped_token" "scoped_token_group" {
  scopes = ["applied-permissions/groups:readers"]
}
```

### Create token with expiry
```hcl
resource "artifactory_scoped_token" "scoped_token_no_expiry" {
  username   = "existing-user"
  expires_in = 7200 // in seconds
}
```

### Creates a refreshable token
```hcl
resource "artifactory_scoped_token" "scoped_token_refreshable" {
  username    = "existing-user"
  refreshable = true
}
```

### Creates an administrator token
```hcl
resource "artifactory_scoped_token" "admin" {
  username = "admin-user"
  scopes   = ["applied-permissions/admin"]
}
```

### Creates a token with an audience
```hcl
resource "artifactory_scoped_token" "audience" {
  username  = "admin-user"
  scopes    = ["applied-permissions/admin"]
  audiences = ["jfrt@*"]
}
```

## Attribute Reference

The following arguments are supported:

* `username` - (Optional) The user name for which this token is created. The username is based on the authenticated user - either from the user of the authenticated token or based on the username (if basic auth was used). The username is then used to set the subject of the token: `<service-id>/users/<username>`. Limited to 255 characters.
* `scopes` - (Optional) The scope of access that the token provides. Access to the REST API is always provided by default. Administrators can set any scope, while non-admin users can only set the scope to a subset of the groups to which they belong.

  The supported scopes include:
  * `applied-permissions/user` - provides user access. If left at the default setting, the token will be created with the user-identity scope, which allows users to identify themselves in the Platform but does not grant any specific access permissions.
  * `applied-permissions/admin` - the scope assigned to admin users.
  * `applied-permissions/groups` - the group to which permissions are assigned by group name (use username to indicate the group name)
  * `system:metrics:r` - for getting the service metrics
  * `system:livelogs:r` - for getting the service livelogsr

  The scope to assign to the token should be provided as a list of scope tokens, limited to 500 characters in total.

  **Resource Permissions**

  From Artifactory 7.38.x, resource permissions scoped tokens are also supported in the REST API. A permission can be represented as a scope token string in the following format: `<resource-type>:<target>[/<sub-resource>]:<actions>`

  Where:
  * `<resource-type>` - one of the permission resource types, from a predefined closed list. Currently, the only resource type that is supported is the `artifact` resource type.
  * `<target>` - the target resource, can be exact name or a pattern
  * `<sub-resource>` - optional, the target sub-resource, can be exact name or a pattern
  * `<actions>` - comma-separated list of action acronyms.

  The actions allowed are <r, w, d, a, m> or any combination of these actions. To allow all actions - use `*`

  Examples:
  * `["applied-permissions/user", "artifact:generic-local:r"]`
  * `["applied-permissions/group", "artifact:generic-local/path:*"]`
  * `["applied-permissions/admin", "system:metrics:r", "artifact:generic-local:*"]`
* `expires_in` - (Optional) The amount of time, in seconds, it would take for the token to expire. An admin shall be able to set whether expiry is mandatory, what is the default expiry, and what is the maximum expiry allowed. Must be non-negative. Default value is based on configuration in `access.config.yaml`. See [API documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-RevokeTokenbyIDrevoketokenbyid) for details.
* `refreshable` - (Optional) Is this token refreshable? Defaults to `false`
* `description` - (Optional) Free text token description. Useful for filtering and managing tokens. Limited to 1024 characters.
* `audiences` - (Optional) A list of the other instances or services that should accept this token identified by their Service-IDs. Limited to total 255 characters. Default to `*@*` if not set. Service ID must begin with `jfrt@`. For instructions to retrieve the Artifactory Service ID see this [documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-GetServiceID).

**Notes:**
- Changing **any** field forces a new resource to be created.

The following additional attributes are exported:

* `access_token` - Returns the access token to authenticate to Artifactory
* `token_type` - Returns the token type
* `subject` - Returns the token type
* `expiry` - Returns the token expiry
* `issued_at` - Returns the token issued at date/time
* `issuer` - Returns the token issuer

## References

- https://www.jfrog.com/confluence/display/JFROG/Access+Tokens
- https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-AccessTokens

## Import

Artifactory **does not** retain scoped tokens and cannot be imported into state.
