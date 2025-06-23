package repository_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRepository_assign_project_key_gh_329(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, map[string]interface{}{
		"name": name,
	})

	localRepositoryWithProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
	 	  project_key = project.{{ .projectKey }}.key
		  project_environments = ["DEV"]
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
				Config: localRepositoryWithProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
		},
	})
}

func TestAccRepository_unassign_project_key_gh_329(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryWithProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
	 	  project_key = project.{{ .projectKey }}.key
		  project_environments = ["DEV"]
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	localRepositoryNoProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
			key = "{{ .name }}"
			project_environments = ["DEV"]
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryWithProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
			{
				Config: localRepositoryNoProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", ""),
				),
			},
		},
	})
}

func TestAccRepository_can_set_two_project_environments_before_7_53_1(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = project.{{ .projectKey }}.key
	 	  project_environments = ["DEV", "PROD"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(util.ProviderMetadata)
					return util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
				},
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.1", "PROD"),
				),
			},
		},
	})
}

func TestAccRepository_invalid_project_environments_before_7_53_1(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = project.{{ .projectKey }}.key
	 	  project_environments = ["Foo"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(util.ProviderMetadata)
					return util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_environment Foo not allowed.*"),
			},
		},
	})
}

func TestAccRepository_invalid_project_environments_after_7_53_1_before_7_107_1(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = project.{{ .projectKey }}.key
	 	  project_environments = ["DEV", "PROD"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(util.ProviderMetadata)
					isSupported, _ := util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
					multiSupported, _ := util.CheckVersion(meta.ArtifactoryVersion, repository.MultipleEnvironmentsSupportedVersion)
					return !(isSupported && !multiSupported), nil
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(fmt.Sprintf(".*for Artifactory versions %s to %s, only one environment.*", repository.CustomProjectEnvironmentSupportedVersion, repository.MultipleEnvironmentsSupportedVersion)),
			},
		},
	})
}

func TestAccRepository_can_set_two_project_environments_after_7_107_1(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			description  = "My Project"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
			max_storage_in_gibibytes   = 10
			block_deployments_on_limit = false
			email_notification         = true
		}

		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = project.{{ .projectKey }}.key
	 	  project_environments = ["DEV", "PROD"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(util.ProviderMetadata)
					multiSupport, err := util.CheckVersion(meta.ArtifactoryVersion, repository.MultipleEnvironmentsSupportedVersion)
					return !multiSupport, err
				},
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.1", "PROD"),
				),
			},
		},
	})
}

func TestAccRepository_invalid_key(t *testing.T) {
	repoName := fmt.Sprintf("test-generic-local-%d", testutil.RandomInt())
	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name": name,
		"key":  "abcd1234567890123456789123456789012345678901234567890123456789012", // 65 chars, too long
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .key }}"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(`.*Attribute key must be 1 - 64 alphanumeric and hyphen characters.*`),
			},
		},
	})
}
