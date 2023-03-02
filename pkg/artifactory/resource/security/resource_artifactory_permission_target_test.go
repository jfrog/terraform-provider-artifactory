package security_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
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

const permissionJustReleaseBundle = `
	resource "artifactory_permission_target" "{{ .permission_name }}" {
		name = "{{ .permission_name }}"
		release_bundle {
			includes_pattern = ["**"]
			repositories = ["release-bundles"]
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
// we can't auto create the repo because of race conditions'
	//resource "artifactory_local_docker_repository" "{{ .repo_name }}" {
	//	key 	     = "{{ .repo_name }}"
	//}

	resource "artifactory_managed_user" "test-user" {
		name     = "terraform"
		email    = "test-user@artifactory-terraform.com"
		password = "Passsw0rd!"
	}

	resource "artifactory_permission_target" "{{ .permission_name }}" {
	  name = "{{ .permission_name }}"

	  repo {
		includes_pattern = ["foo/**"]
		excludes_pattern = ["bar/**"]
		repositories     = ["{{ .repo_name }}"]

		actions {
			users {
				name        = artifactory_managed_user.test-user.name
				permissions = ["read", "write", "annotate", "delete"]
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
				name        = artifactory_managed_user.test-user.name
				permissions = ["read", "write", "manage", "annotate", "delete"]
			}

		  groups {
			name        = "readers"
			permissions = ["read"]
		  }
		}
	  }

	  release_bundle {
		includes_pattern = ["foo/**"]
		excludes_pattern = ["bar/**"]
		repositories     = ["release-bundles"]

		actions {
			users {
				name        = artifactory_managed_user.test-user.name
				permissions = ["read", "write", "managedXrayMeta", "distribute"]
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

func TestAccPermissionTarget_GitHubIssue126(t *testing.T) {
	_, permFqrn, permName := test.MkNames("test-perm", "artifactory_permission_target")
	_, _, repoName := test.MkNames("test-perm-repo", "artifactory_local_generic_repository")
	_, _, username := test.MkNames("artifactory_user", "artifactory_user")
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
		  password 					 = "Passw0rd!"
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
	foo := util.ExecuteTemplate(permFqrn, testConfig, variables)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
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
			{
				ResourceName:      permFqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(permName, "name"),
			},
		},
	})
}

func TestAccPermissionTarget_full(t *testing.T) {
	_, permFqrn, permName := test.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       "example-repo-local",
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, permissionFull, tempStruct),
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
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.excludes_pattern.#", "1"),
				),
			},
			{
				ResourceName:      permFqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(permName, "name"),
			},
		},
	})
}

func TestAccPermissionTarget_user_permissions(t *testing.T) {
	_, permFqrn, permName := test.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       "example-repo-local",
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, permissionFull, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),

					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.0.permissions.#", "4"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "read"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "write"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "annotate"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "delete"),

					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.0.permissions.#", "5"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "read"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "write"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "manage"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "annotate"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "delete"),

					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.0.permissions.#", "4"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "release_bundle.0.actions.0.users.0.permissions.*", "read"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "release_bundle.0.actions.0.users.0.permissions.*", "write"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "release_bundle.0.actions.0.users.0.permissions.*", "managedXrayMeta"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "release_bundle.0.actions.0.users.0.permissions.*", "distribute"),
				),
			},
			{
				ResourceName:      permFqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(permName, "name"),
			},
		},
	})
}

func TestAccPermissionTarget_addBuild(t *testing.T) {
	_, permFqrn, permName := test.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       "example-repo-local", // because of race conditions in artifactory, this repo must first exist
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, permissionNoIncludes, tempStruct),
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
				Config: util.ExecuteTemplate(permFqrn, permissionJustBuild, tempStruct),
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
				Config: util.ExecuteTemplate(permFqrn, permissionJustReleaseBundle, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "build.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.groups.#", "0"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.excludes_pattern.#", "0"),
				),
			},
			{
				Config: util.ExecuteTemplate(permFqrn, permissionFull, tempStruct),
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
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.excludes_pattern.#", "1"),
				),
			},
			{
				ResourceName:      permFqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(permName, "name"),
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

			exists, _ := security.PermTargetExists(rs.Primary.ID, acctest.Provider.Meta())
			if !exists {
				return nil
			}
			return fmt.Errorf("error: Permission targets %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
