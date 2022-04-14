package artifactory

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const permissionNoIncludes = `
	//resource "artifactory_local_docker_repository" "{{ .repo_name }}" {
	//	key 	     = "{{ .repo_name }}"
	//}
	resource "artifactory_permission_target" "{{ .permission_name }}" {
		name = "{{ .permission_name }}"
		repo {
			repositories = ["{{ .repo_name }}"]
			actions {
				users {
				  name = "anonymous"
				  permissions = ["read", "write"]
				}
			}
		}
     //depends_on = [artifactory_local_docker_repository.{{ .repo_name }}]

	}
`

const permissionJustBuild = `
	//resource "artifactory_local_docker_repository" "{{ .repo_name }}" {
	//	key 	     = "{{ .repo_name }}"
	//}
	resource "artifactory_permission_target" "{{ .permission_name }}" {
		name = "{{ .permission_name }}"
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
		//depends_on = [artifactory_local_docker_repository.{{ .repo_name }}]

	}
`

const permissionFull = `
// we can't auto create the repo because of race conditions'
	//resource "artifactory_local_docker_repository" "{{ .repo_name }}" {
	//	key 	     = "{{ .repo_name }}"
	//}

	resource "artifactory_permission_target" "{{ .permission_name }}" {
	  name = "{{ .permission_name }}"

	  repo {
		includes_pattern = ["foo/**"]
		excludes_pattern = ["bar/**"]
		repositories     = ["{{ .repo_name }}"]

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
     //depends_on = [artifactory_local_docker_repository.{{ .repo_name }}]
	}
`

func TestTestAccPermissionTarget_GitHubIssue126(test *testing.T) {
	_, permFqrn, permName := mkNames("test-perm", "artifactory_permission_target")
	_, _, repoName := mkNames("test-perm-repo", "artifactory_local_generic_repository")
	_, _, username := mkNames("artifactory_user", "artifactory_user")
	testConfig := `
		resource "artifactory_local_generic_repository" "{{ .repo_name }}" {
		  key             = "{{ .repo_name }}"
		  repo_layout_ref = "simple-default"
		}

		resource "artifactory_user" "{{ .username }}" {
		  name                       = "{{ .username }}"
		  email                      = "example@example.com"
		  groups                     = ["readers"]
		  admin                      = false
		  disable_ui_access          = true
		  internal_password_disabled = true
		  password 					 = "Password1"
		}

		resource "artifactory_permission_target" "{{ .perm_name }}" {
		  name = "{{ .perm_name }}"
		  repo {
			includes_pattern = ["**"]
			repositories = [
			  "{{ .repo_name }}"
			]
			actions {
			  users {
				name        = artifactory_user.{{ .username }}.name
				permissions = ["annotate", "read", "write", "delete"]
			  }
			}
		  }
		}`
	variables := map[string]string{
		"perm_name": permName,
		"username":  username,
		"repo_name": repoName,
	}
	foo := executeTemplate(permFqrn, testConfig, variables)
	resource.Test(test, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(test) },
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: foo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.0.permissions.#", "4"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.repositories.#", "1"),
				),
			},
		},
	})
}

func TestAccPermissionTarget_full(test *testing.T) {
	_, permFqrn, permName := mkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       "example-repo-local",
		"permission_name": permName,
	}

	resource.Test(test, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(test) },
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(permFqrn, permissionFull, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.excludes_pattern.#", "1"),
				),
			},
		},
	})
}

func TestAccPermissionTarget_addBuild(t *testing.T) {
	_, permFqrn, permName := mkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       "example-repo-local", // because of race conditions in artifactory, this repo must first exist
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(permFqrn, permissionNoIncludes, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.includes_pattern.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.excludes_pattern.#", "0"),
				),
			},
			{
				Config: executeTemplate(permFqrn, permissionJustBuild, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.excludes_pattern.#", "0"),
				),
			},
			{
				Config: executeTemplate(permFqrn, permissionFull, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.excludes_pattern.#", "1"),
				),
			},
		},
	})
}

func testPermissionTargetCheckDestroy(id ...string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, fqrn := range id {
			rs, ok := s.RootModule().Resources[fqrn]

			if !ok {
				return fmt.Errorf("err: Resource id[%s] not found", id)
			}
			provider, _ := testAccProviders["artifactory"]()
			exists, _ := permTargetExists(rs.Primary.ID, provider.Meta())
			if !exists {
				return nil
			}
			return fmt.Errorf("error: Permission targets %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
