package artifactory

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const permissionNoIncludes = `
resource "artifactory_permission_target" "test-perm" {
	name = "test-perm"
	repo {
		repositories = ["example-repo-local"]
		actions {
			users {
			  name = "anonymous"
			  permissions = ["read", "write"]
			}
		}
	}
}
`

const permissionJustBuild = `
resource "artifactory_permission_target" "test-perm" {
	name = "test-perm"
	build {
		includes_pattern = ["**"]
		repositories = ["artifactory-build-info"]
		actions {
			users {
				name = "anonymous"
				permissions = ["read", "write"]
			}
		}
	}
}
`

const permissionFull = `
resource "artifactory_permission_target" "test-perm" {
  name = "test-perm"

  repo {
    includes_pattern = ["foo/**"]
    excludes_pattern = ["bar/**"]
    repositories     = ["example-repo-local"]

    actions {
      users {
		name        = "anonymous"
		permissions = ["read", "write"]
	  }

      groups {
        name        = "readers"
        permissions = ["read"]
      }
    }
  }

  build {
    includes_pattern = ["foo/**"]
    excludes_pattern = ["bar/**"]
    repositories     = ["artifactory-build-info"]

    actions {
      users {
        name        = "anonymous"
        permissions = ["read", "write"]
      }
      
      
      groups {
        name        = "readers"
        permissions = ["read"]
      }
    }
  }
}
`

func TestAccPermissionTarget_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testPermissionTargetCheckDestroy("artifactory_permission_target.test-perm"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: permissionFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "name", "test-perm"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.excludes_pattern.#", "1"),
				),
			},
		},
	})
}

func TestAccPermissionTarget_addBuild(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testPermissionTargetCheckDestroy("artifactory_permission_target.test-perm"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: permissionNoIncludes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "name", "test-perm"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.groups.#", "0"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.includes_pattern.#", "0"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.excludes_pattern.#", "0"),
				),
			},
			{
				Config: permissionJustBuild,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "name", "test-perm"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.#", "0"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.groups.#", "0"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.excludes_pattern.#", "0"),
				),
			},
			{
				Config: permissionFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "name", "test-perm"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "build.0.excludes_pattern.#", "1"),
				),
			},
		},
	})
}

func testPermissionTargetCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		exists, err := client.V2.Security.HasPermissionTarget(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else if !exists {
			return nil
		} else {
			return fmt.Errorf("error: Permission targets %s still exists", rs.Primary.ID)
		}
	}
}
