### Create a new Artifactory scoped token for an existing user

resource "artifactory_scoped_token" "scoped_token" {
  username = "existing-user"
}

### **Note:** This assumes that the user `existing-user` has already been created in Artifactory by different means, i.e. manually or in a separate terraform apply.

### Create a new Artifactory user and scoped token
resource "artifactory_user" "new_user" {
  name   = "new_user"
  email  = "new_user@somewhere.com"
  groups = ["readers"]
}

resource "artifactory_scoped_token" "scoped_token_user" {
  username = artifactory_user.new_user.name
}

### Creates a new token for groups
resource "artifactory_scoped_token" "scoped_token_group" {
  scopes = ["applied-permissions/groups:readers"]
}

### Create token with expiry
resource "artifactory_scoped_token" "scoped_token_no_expiry" {
  username   = "existing-user"
  expires_in = 7200 // in seconds
}

### Creates a refreshable token
resource "artifactory_scoped_token" "scoped_token_refreshable" {
  username    = "existing-user"
  refreshable = true
}

### Creates an administrator token
resource "artifactory_scoped_token" "admin" {
  username = "admin-user"
  scopes   = ["applied-permissions/admin"]
}

### Creates a token with an audience
resource "artifactory_scoped_token" "audience" {
  username  = "admin-user"
  scopes    = ["applied-permissions/admin"]
  audiences = ["jfrt@*"]
}