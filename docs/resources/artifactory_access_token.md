# Artifactory Access Token Resource

Provides an Artifactory Access Token resource. This can be used to create and manage Artifactory Access Tokens.

~> **Note:** Access Tokens will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).


## Example Usages

### Create a new Artifactory Access Token for an existing user

```hcl
resource "artifactory_access_token" "exising_user" {
  username          = "existing-user"
  end_date_relative = "5m"
}
```

~> **Note:** This assumes that the user `existing-user` has already been created in Artifactory by different means, (e.g., manually, or in a separate `terraform apply`).

### Create a new Artifactory User and Access token

```hcl
resource "artifactory_user" "new_user" {
  name   = "new_user"
  email  = "new_user@somewhere.com"

  groups = [
    "readers",
  ]
}

resource "artifactory_access_token" "new_user" {
  username          = artifactory_user.new_user.name
  end_date_relative = "5m"
}
```

### Creates a new token for groups

This creates an ephemeral user called `temporary-user`.

```hcl
resource "artifactory_access_token" "temporary_user" {
  username          = "temporary-user"
  end_date_relative = "1h"

  groups = [
    "readers",
  ]
}
```

### Create token with no expiry

```hcl
resource "artifactory_access_token" "no_expiry" {
  username          = "existing-user"
  end_date_relative = "0s"
}
```

### Creates a refreshable token

```hcl
resource "artifactory_access_token" "refreshable" {
  username          = "refreshable"
  end_date_relative = "1m"

  refreshable = true

  groups = [
    "readers",
  ]
}
```

### Creates an administrator token

```hcl
resource "artifactory_access_token" "admin" {
  username          = "admin"
  end_date_relative = "1m"

  admin_token {
    instance_id = "jfrt@<instance id>"
  }
}
```

### Creates a token with an audience

```hcl
resource "artifactory_access_token" "audience" {
  username          = "audience"
  end_date_relative = "1m"

  audience    = "jfrt@*"
  refreshable = true
}
```

### Creates a token with a fixed end date

```hcl
resource "artifactory_access_token" "fixeddate" {
  username = "fixeddate"
  end_date = "2018-01-01T01:02:03Z"

  groups = [
    "readers",
  ]
}
```

### Rotate token after it expires

This example will generate a token that will expire in 1 hour. If `terraform apply` is run before 1 hour has passed, nothing changes. Once an hour has passed, `terraform apply` will generate a new token.

```hcl
resource "time_rotating" "now_plus_1_hours" {
  rotation_hours = "1"
}

resource "artifactory_access_token" "rotating" {
  username = "rotating"

  # the end_date is set to now + 1 hours
  end_date = time_rotating.now_plus_1_hour.rotation_rfc3339

  groups = [
    "readers",
  ]
}
```

### Rotate token each `terraform apply`

This example will generate a token that will expire in 1 hour. If `terraform apply` is run before 1 hour has passed, a new token is generated with an expiry of 1 hour.

```hcl
resource "time_rotating" "now_plus_1_hours" {
  triggers = {
    "key" = timestamp()
  }

  rotation_hours = "1"
}

resource "artifactory_access_token" "rotating" {
  username = "rotating"

  # the end_date is set to now + 1 hours
  end_date = time_rotating.now_plus_1_hour.rotation_rfc3339

  groups = [
    "readers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The username or subject for the token. A non-admin can only specify their own username. Admins can specify any existing username, or a new name for a temporary token. Temporary tokens require `groups` to be set.

* One of `end_date` or `end_date_relative` must be set.

    * `end_date` - (Optional) The end date which the token is valid until, formatted as an [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) date string (e.g. `2018-01-01T01:02:03Z`).

    * `end_date_relative` - (Optional) A relative duration for which the token is valid. For example `240h` (10 days) or `2400h30m`. Valid time units are the same as Golang’s [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) method: `s`, `m`, `h`.

* `groups` - (Optional) List of groups. The token is granted access based on the permissions of the groups. Specify `["*"]` for all groups that the user belongs to. `groups` cannot be specified with `admin_token`.

* `admin_token` - (Optional) Specify the `instance_id` in this block to grant this token admin privileges. This can only be created when the authenticated user is an admin. `admin_token` cannot be specified with `groups`.

* `refreshable` - (Optional) Should this token refreshable? A value of `true` means that the token is refreshable and the response will include a refresh token. A value of `false` means that the token is _not_ refreshable and the response will not include a refresh token value. The default value is `false`.

* `audience` - (Optional) A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs. You may set `jfrt@*` so the token to be accepted by all Artifactory instances.

    The `refreshable` attribute must be set to `true` in order to set the `audience`. [How to retrieve the Artifactory Service ID](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-GetServiceID).

### Notes

* Changing **any** field forces a new resource to be created.

* Although you can create a refreshable token by setting `refreshable` to true, the resource does **not** implement a token refresh on subsequent executions of Terraform.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `access_token` - Returns the access token for authenciating with Artifactory.

* `refresh_token` - Returns the refresh token when `refreshable` is set to `true`, or an empty string when `refreshable` is set to `false`.

## References

* https://www.jfrog.com/confluence/display/ACC1X/Access+Tokens
* https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateToken

## Import

Artifactory **does not** retain access tokens and cannot be imported into state.
