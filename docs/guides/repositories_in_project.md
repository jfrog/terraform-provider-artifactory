---
page_title: "Adding repositories to the project"
---

The guide provides information and the example on how to add repositories to the project. 

## Artifactory behavior

The attribute `project_environments` (`environments` in the API call) is ignored by Artifactory, if the repository is not assigned to an existing project.
That attribute can only be set to the repository if it's assigned to the project already. 
The project can't be created with the list of non-existing repositories, and the repository can't be assigned to non-existing project.
Thus, if the project and the repository/repositories are created at the same time in one Terraform configuration, we have a state drift 
for the `project_environments` attribute.

This is happening, because the repositories need to be created first, then the project with the list of repositories gets 
created. Since `project_environments` attribute is ignored in the first step, we will only have this attribute in the Terraform state, and
not in the actual repository properties.

On the next step, when the repo gets assigned to the project by the project resource, the default value `DEV` is assigned to 
repositories' `project_environments` attribute. If the desired value on the first step was `DEV`, then the values match and no state
drift occurs. But if the desired value was `PROD`, we will get an error message when updating the config by `terraform plan`/`terraform apply`.

```
  # artifactory_local_docker_v2_repository.docker-v2-local will be updated in-place
  ~ resource "artifactory_local_docker_v2_repository" "docker-v2-local" {
        id                       = "myproj-docker-v2-local"
      ~ project_environments     = [
          - "DEV",
          + "PROD",
        ]
    } 
```

## Workaround

~> In the Project provider documentation, we strongly recommend using the `repos` attribute to manage the list of repositories.
Do not use `project_key` attribute of the repository resource.

Unfortunately, we can't fix the behavior described above right now. The workaround is simply to run `terraform apply` twice. 
When the user applies the configuration second time, the repository is already assigned to the project, and `project_environments` 
attribute won't be ignored.

## Full HCL example

```hcl
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "6.21.4"
    }
    project = {
      source  = "registry.terraform.io/jfrog/project"
      version = "1.1.1"
    }
  }
}

provider "artifactory" {
  // supply ARTIFACTORY_ACCESS_TOKEN / JFROG_ACCESS_TOKEN / ARTIFACTORY_API_KEY and ARTIFACTORY_URL / JFROG_URL as env vars
}

resource "artifactory_local_docker_v2_repository" "docker-v2-local" {
  key                   = "myproj-docker-v2-local"
  tag_retention         = 3
  max_unique_tags       = 5
  project_environments  = ["PROD"]

  lifecycle {
    ignore_changes = [
      project_key
    ]
  }
}

resource "project" "myproject" {
  key          = "myproj"
  display_name = "My Project"
  description  = "My Project"
  admin_privileges {
    manage_members   = true
    manage_resources = true
    index_resources  = true
  }
  max_storage_in_gibibytes   = 10
  block_deployments_on_limit = false
  email_notification         = true

  repos = ["myproj-docker-v2-local"]

  depends_on = [ artifactory_local_docker_v2_repository.docker-v2-local ]
}
```
~> Apply `lifecycle.ignore_changes` to `project_key` attribute, otherwise it will be removed from the repository, 
which means it will be unassigned from the project on the configuration update.
