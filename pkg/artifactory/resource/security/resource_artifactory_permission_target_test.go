package security_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

const permissionNoIncludes = `
	resource "artifactory_local_docker_v2_repository" "{{ .repo_name }}" {
		key = "{{ .repo_name }}"
	}

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
		depends_on = [artifactory_local_docker_v2_repository.{{ .repo_name }}]
	}
`

const permissionJustBuild = `
	resource "artifactory_local_docker_v2_repository" "{{ .repo_name }}" {
		key = "{{ .repo_name }}"
	}
	
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

		depends_on = [artifactory_local_docker_v2_repository.{{ .repo_name }}]
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
	resource "artifactory_local_docker_v2_repository" "{{ .repo_name }}" {
		key = "{{ .repo_name }}"
	}

	resource "artifactory_managed_user" "test-user" {
		name     = "terraform"
		email    = "test-user@artifactory-terraform.com"
		password = "Passsw0rd!12"
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

	  depends_on = [artifactory_local_docker_v2_repository.{{ .repo_name }}]
	}
`

const testLength = `
	// we can't auto create the repo because of race conditions'
	resource "artifactory_local_docker_v2_repository" "{{ .repo_name }}" {
		key = "{{ .repo_name }}"
	}

	resource "artifactory_permission_target" "{{ .permission_name }}" {
	  name = "{{ .permission_name }}"

	  repo {
		includes_pattern = ["foo/**"]
		repositories     = ["{{ .repo_name }}"]
	  }
	  depends_on = [artifactory_local_docker_v2_repository.{{ .repo_name }}]
	}
`

func TestAccPermissionTarget_noActions(t *testing.T) {
	repoName := fmt.Sprintf("test-local-docker-%d", rand.Int())
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       repoName,
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, testLength, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.#", "0"),
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

func TestAccPermissionTarget_MigrateFromFrameworkBackToSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-perm", "artifactory_permission_target")
	repoName := fmt.Sprintf("test-local-docker-%d", rand.Int())

	data := map[string]string{
		"repo_name":       repoName,
		"permission_name": name,
	}

	config := util.ExecuteTemplate(fqrn, permissionFull, data)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "9.7.3",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "repo.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.#", "1"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func TestAccPermissionTarget_GitHubIssue126(t *testing.T) {
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")
	_, _, repoName := testutil.MkNames("test-perm-repo", "artifactory_local_generic_repository")
	_, _, username := testutil.MkNames("artifactory_managed_user", "artifactory_managed_user")
	testConfig := `
		resource "artifactory_local_generic_repository" "{{ .repo_name }}" {
		  key             = "{{ .repo_name }}"
		  repo_layout_ref = "simple-default"
		}

		resource "artifactory_managed_user" "{{ .username }}" {
		  name                       = "{{ .username }}"
		  email                      = "example@example.com"
		  groups                     = ["readers"]
		  admin                      = false
		  disable_ui_access          = true
		  internal_password_disabled = false
		  password 					 = "Passw0rd!123"
		}

		resource "artifactory_permission_target" "{{ .perm_name }}" {
		  name = "{{ .perm_name }}"
		  repo {
			includes_pattern = ["**"]
			repositories     = ["{{ .repo_name }}"]

			actions {
			  users {
				name        = artifactory_managed_user.{{ .username }}.name
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, testConfig, variables),
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
	repoName := fmt.Sprintf("test-local-docker-%d", rand.Int())
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       repoName,
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
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
	repoName := fmt.Sprintf("test-local-docker-%d", rand.Int())
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       repoName,
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(permFqrn, permissionFull, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(permFqrn, "name", permName),

					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.0.name", "terraform"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.users.0.permissions.#", "4"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "read"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "write"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "annotate"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.users.0.permissions.*", "delete"),

					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.0.name", "readers"),
					resource.TestCheckResourceAttr(permFqrn, "repo.0.actions.0.groups.0.permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "repo.0.actions.0.groups.0.permissions.*", "read"),

					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.0.name", "terraform"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.users.0.permissions.#", "5"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "read"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "write"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "manage"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "annotate"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.users.0.permissions.*", "delete"),

					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.0.name", "readers"),
					resource.TestCheckResourceAttr(permFqrn, "build.0.actions.0.groups.0.permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(permFqrn, "build.0.actions.0.groups.0.permissions.*", "read"),

					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(permFqrn, "release_bundle.0.actions.0.users.0.name", "terraform"),
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
	repoName := fmt.Sprintf("test-local-docker-%d", rand.Int())
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")

	tempStruct := map[string]string{
		"repo_name":       repoName,
		"permission_name": permName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
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

func TestAccPermissionTarget_MissingRepositories(t *testing.T) {
	_, permFqrn, permName := testutil.MkNames("test-perm", "artifactory_permission_target")
	_, _, repoName := testutil.MkNames("test-perm-repo", "artifactory_local_generic_repository")
	_, _, username := testutil.MkNames("artifactory_managed_user", "artifactory_managed_user")
	testConfig := `
		resource "artifactory_managed_user" "{{ .username }}" {
		  name                       = "{{ .username }}"
		  email                      = "example@example.com"
		  groups                     = ["readers"]
		  admin                      = false
		  disable_ui_access          = true
		  internal_password_disabled = true
		  password 					 = "Passw0rd!123"
		}

		resource "artifactory_permission_target" "{{ .perm_name }}" {
		  name = "{{ .perm_name }}"
		  repo {
			includes_pattern = ["**"]

			actions {
			  users {
				name        = artifactory_managed_user.{{ .username }}.name
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testPermissionTargetCheckDestroy(permFqrn),
		Steps: []resource.TestStep{
			{
				Config:      util.ExecuteTemplate(permFqrn, testConfig, variables),
				ExpectError: regexp.MustCompile(".*Missing required argument.*"),
			},
		},
	})
}

func testPermissionTargetCheckDestroy(id ...string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		var err error

		for _, fqrn := range id {
			rs, ok := s.RootModule().Resources[fqrn]

			if !ok {
				err = fmt.Errorf("err: Resource id[%s] not found", id)
				break
			}

			exists, _ := security.PermTargetExists(rs.Primary.ID, acctest.Provider.Meta())
			if !exists {
				continue
			}

			err = fmt.Errorf("error: Permission targets %s still exists", rs.Primary.ID)
			break
		}

		return err
	}
}
