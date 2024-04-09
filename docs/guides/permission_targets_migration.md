---
page_title: "Migrating from Artifactory Permission Target resource to Platform Permission resource"
---

Artifactory version 7.72.0 introduced new [Access Permission API](https://jfrog.com/help/r/jfrog-rest-apis/permissions) which replaces the Artifactory's [Permission Target API](https://jfrog.com/help/r/jfrog-rest-apis/permission-targets). A new resource [`platform_permission`](https://registry.terraform.io/providers/jfrog/platform/latest/docs/resources/permission) is added to the Platform Terraform provider to support the new Access Permission.

While `artifactory_permission_target` continues to be supported and can still be used to manage permissions, we recommend only using `platform_permission` for any new permission resources.

~>Only `platform_permission` supports managing permission for Release Lifecycle (a.k.a. Destination), and Pipeline Source.

->The Access Permissions API are applicable for the next-generation permissions model and fully backwards compatible with the legacy permissions model. You can continue to use the previous APIs for your workflows.

The schema differences between `artifactory_permission_target` and `platform_permission` are minor and mainly centered around how the permission targets are defined.

Using the following configuration with `artifactory_permission_target`:

```terraform
resource "artifactory_permission_target" "my-permission" {
  name = "my-permission-name"

  repo {
    includes_pattern = ["foo/**"]
    excludes_pattern = ["bar/**"]
    repositories     = ["example-repo-local"]

    actions {
      users {
        name        = "anonymous"
        permissions = ["read", "write"]
      }

      users {
        name        = "user1"
        permissions = ["read", "write"]
      }

      groups {
        name        = "readers"
        permissions = ["read"]
      }

      groups {
        name        = "dev"
        permissions = ["read", "write"]
      }
    }
  }

  build {
    includes_pattern = ["**"]
    repositories     = ["artifactory-build-info"]

    actions {
      users {
        name        = "anonymous"
        permissions = ["read"]
      }

      users {
        name        = "user1"
        permissions = ["read", "write"]
      }
    }
  }

  release_bundle {
    includes_pattern = ["**"]
    repositories     = ["release-bundles"]

    actions {
      users {
        name         = "anonymous"
        permissions  = ["read"]
      }

      groups {
        name        = "readers"
        permissions = ["read"]
      }
    }
  }
}
```

The corresponding configuration for `platform_permission` will be:

```terraform
resource "platform_permission" "my-permission" {
  name = "my-permission-name"

  artifact = {
    targets = [
      {
        name = "example-repo-local"
        include_patterns = ["foo/**"]
        exclude_patterns = ["bar/**]
      }
    ]

    actions = {
      users = [
        {
          name = "anonymous"
          permissions = ["READ", "WRITE"]
        },
        {
          name = "user1"
          permissions = ["READ", "WRITE"]
        },
      ]

      groups = [
        {
          name = "readers"
          permissions = ["READ"]
        },
        {
          name = "dev"
          permissions = ["READ", "WRITE"]
        },
      ]
    }
  }

  build = {
    targets = [
      {
        name = "artifactory-build-info"
        include_patterns = ["**"]
      }
    ]

    actions = {
      users = [
        {
          name = "anonymous"
          permissions = ["READ"]
        },
        {
          name = "user1"
          permissions = ["READ", "WRITE"]
        },
      ]
    }
  }

  release_bundle = {
    targets = [
      {
        name = "release-bundle"
        include_patterns = ["**"]
      }
    ]

    actions = {
      users = [
        {
          name = "anonymous"
          permissions = ["READ"]
        }
      ]

      groups = [
        {
          name = "readers"
          permissions = ["READ"]
        }
      ]
    }
  }
}
```

## Schema Differences

| Type            |   `artifactory_permission_target`                                                                                      | `platform_permission`                                                                                                                                                                   |
|-----------------|------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Attribute name  | <pre>repo {<br>  ...<br>}</pre>                                                                                        | <pre>artifact = {<br>  ...<br>}</pre>                                                                                                                                                   |
| Block           | `repo`, `build`, or `release_bundle`<br><pre>repo {<br>  repositories = ["example-repo-local"]<br>}</pre>              | List of  `artifact.targets`,  `build.targets`,  `release_bundle.targets`<br><pre>artifact = {<br>  targets = [<br>    {<br>      name = "example-repo-local"<br>    }<br>  ]<br>}</pre> |
| Attribute name  | `repo.repositories`                                                                                                    | `artifact.targets.name`                                                                                                                                                                 |
| Attribute name  | `includes_pattern`                                                                                                     | `include_patterns`                                                                                                                                                                      |
| Attribute name  | `excludes_pattern`                                                                                                     | `exclude_patterns`                                                                                                                                                                      |
| Block           | Multiple of `actions.users`<br><pre>actions {<br>  users {<br>    ...<br>  }<br>  users {<br>    ...<br>  }<br>}</pre> | List of  `actions.users`<br><pre>actions = {<br>  users = [<br>    {<br>      ...<br>    },<br>    {<br>      ...<br>    }<br>  ]<br>}</pre>                                            |
| Attribute value | Lower case  <pre>permissions = ["read", "write"]</pre>                                                                 | UPPER CASE <pre>permissions = ["READ", "WRITE"]</pre>                                                                                                                                   |
