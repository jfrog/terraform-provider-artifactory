# Artifactory Access Token Resource

Provides an Artifactory access token resource. This can be used to create and revoke access tokens

~> **Note:** Access and refresh tokens will be stored in the raw state as plain-text. [Read more about sensitive data in
state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```hcl
# Create a new Artifactory access token as the configured user
resource "artifactory_access_token" "my_token" {
  username = "my-token"
  scope    = "api:*"
}

output "my_access_token" {
  value = artifactory_access_token.my_token.access_token
}

output "my_refresh_token" {
  value = artifactory_access_token.my_token.refresh_token
}
```

## Attribute Reference

The following attributes are exported:

* `username` - (Required) The user name for which this token is created
* `scope` - (Required) The scope to assign to the token provided as a space-separated list of scope tokens
* `expires_in` - (Optional) The time in seconds for which the token will be valid
* `refreshable` - (Optional) If true, this token is refreshable and the refresh token can be used to replace it with a new token once it expires
* `audience` - (Optional) A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory [Service IDs](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-GetServiceID) as obtained from the Get Service ID endpoint

See the Artifactory REST API [documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateToken) for more details
