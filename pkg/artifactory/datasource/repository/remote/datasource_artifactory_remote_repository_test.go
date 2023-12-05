package remote_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccDataSourceRemoteAllBasicPackageTypes(t *testing.T) {
	for _, packageType := range remote.PackageTypesLikeBasic {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkTestCase(packageType, t))
		})
	}
}

func mkTestCase(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("remote-%s-%d", packageType, testutil.RandomInt())
	resourceName := fmt.Sprintf("data.artifactory_remote_%s_repository.%s", packageType, name)

	params := map[string]interface{}{
		"packageType": packageType,
		"name":        name,
	}
	config := utilsdk.ExecuteTemplate("TestAccRemoteRepository", `
		resource "artifactory_remote_{{ .packageType }}_repository" "{{ .name }}" {
		    key         = "{{ .name }}"
		    description = "Test repo for {{ .name }}"
		    notes       = "Test repo for {{ .name }}"
		    url         = "http://tempurl.org"
		}

		data "artifactory_remote_{{ .packageType }}_repository" "{{ .name }}" {
		    key = artifactory_remote_{{ .packageType }}_repository.{{ .name }}.id
		}
	`, params)

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", packageType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "url", "http://tempurl.org"),
				),
			},
		},
	}
}

func TestAccDataSourceRemoteBowerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("bower-remote", "data.artifactory_remote_bower_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteBowerRepository",
		`resource "artifactory_remote_bower_repository" "{{ .name }}" {
		    key = "{{ .name }}"
		    url = "http://tempurl.org"
		}

		data "artifactory_remote_bower_repository" "{{ .name }}" {
		    key = artifactory_remote_bower_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "bower"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "bower-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "bower_registry_url", "https://registry.bower.io"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteCargoRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cargo-remote", "data.artifactory_remote_cargo_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteCargoRepository",
		`resource "artifactory_remote_cargo_repository" "{{ .name }}" {
		    key                 = "{{ .name }}"
		    url                 = "http://tempurl.org"
		    git_registry_url    = "http://tempurl.org"
		    anonymous_access    = true
		    enable_sparse_index = true
		}

		data "artifactory_remote_cargo_repository" "{{ .name }}" {
		    key = artifactory_remote_cargo_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "cargo"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "git_registry_url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", "true"),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteCocoaPodsRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cocoapods-remote", "data.artifactory_remote_cocoapods_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteCocoaPodsRepository",
		`resource "artifactory_remote_cocoapods_repository" "{{ .name }}" {
		    key                 = "{{ .name }}"
		    url                 = "http://tempurl.org"
		    pods_specs_repo_url = "http://tempurl.org"
		}

		data "artifactory_remote_cocoapods_repository" "{{ .name }}" {
		    key = artifactory_remote_cocoapods_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "cocoapods"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "pods_specs_repo_url", "http://tempurl.org"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteComposerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("composer-remote", "data.artifactory_remote_composer_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteComposerRepository",
		`resource "artifactory_remote_composer_repository" "{{ .name }}" {
		    key                   = "{{ .name }}"
		    url                   = "http://tempurl.org"
		    composer_registry_url = "http://tempurl.org"
		}

		data "artifactory_remote_composer_repository" "{{ .name }}" {
		    key = artifactory_remote_composer_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "composer"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "composer-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "composer_registry_url", "http://tempurl.org"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteConanRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("conan-remote", "data.artifactory_remote_conan_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteConanRepository",
		`resource "artifactory_remote_conan_repository" "{{ .name }}" {
		    key                        = "{{ .name }}"
		    url                        = "http://tempurl.org"
		    force_conan_authentication = true
		}

		data "artifactory_remote_conan_repository" "{{ .name }}" {
		    key = artifactory_remote_conan_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "conan"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "conan-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteDockerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-remote", "data.artifactory_remote_docker_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteConanRepository",
		`resource "artifactory_remote_docker_repository" "{{ .name }}" {
		    key                            = "{{ .name }}"
		    url                            = "http://tempurl.org"
		    external_dependencies_enabled  = true
		    enable_token_authentication    = true
		    block_pushing_schema1          = true
		    external_dependencies_patterns = ["*foo"]
			curated                        = false
		}

		data "artifactory_remote_docker_repository" "{{ .name }}" {
		    key = artifactory_remote_docker_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "docker"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "true"),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "*foo"),
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteGenericRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("generic-remote", "data.artifactory_remote_generic_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteGenericRepository",
		`resource "artifactory_remote_generic_repository" "{{ .name }}" {
		    key = "{{ .name }}"
		    url = "http://tempurl.org"
		}

		data "artifactory_remote_generic_repository" "{{ .name }}" {
		    key = artifactory_remote_generic_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteGoRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("go-remote", "data.artifactory_remote_go_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteGoRepository",
		`resource "artifactory_remote_go_repository" "{{ .name }}" {
		    key              = "{{ .name }}"
		    url              = "http://tempurl.org"
		    vcs_git_provider = "GITHUB"
		}

		data "artifactory_remote_go_repository" "{{ .name }}" {
		    key = artifactory_remote_go_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "go"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "go-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "vcs_git_provider", "GITHUB"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteHelmRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("helm-remote", "data.artifactory_remote_helm_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteHelmRepository",
		`resource "artifactory_remote_helm_repository" "{{ .name }}" {
		    key                            = "{{ .name }}"
		    url                            = "http://tempurl.org"
		    helm_charts_base_url           = "http://tempurl.org"
		    external_dependencies_enabled  = true
		    external_dependencies_patterns = ["*foo"]
		}

		data "artifactory_remote_helm_repository" "{{ .name }}" {
		    key = artifactory_remote_helm_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "helm"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "helm_charts_base_url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "*foo"),
				),
			},
		},
	})
}

var commonJavaParams = func() map[string]interface{} {
	return map[string]interface{}{
		"name":                             "",
		"url":                              "http://tempurl.org",
		"fetch_jars_eagerly":               testutil.RandBool(),
		"fetch_sources_eagerly":            testutil.RandBool(),
		"remote_repo_checksum_policy_type": testutil.RandSelect("generate-if-absent", "fail", "ignore-and-generate", "pass-thru"),
		"handle_releases":                  testutil.RandBool(),
		"handle_snapshots":                 testutil.RandBool(),
		"suppress_pom_consistency_checks":  testutil.RandBool(),
		"reject_invalid_jars":              testutil.RandBool(),
	}
}

const javaRepositoryBasic = `
resource "{{ .resource_name }}" "{{ .name }}" {
    key                              = "{{ .name }}"
    url                              = "{{ .url }}"
    fetch_jars_eagerly               = {{ .fetch_jars_eagerly }}
    fetch_sources_eagerly            = {{ .fetch_sources_eagerly }}
    remote_repo_checksum_policy_type = "{{ .remote_repo_checksum_policy_type }}"
    handle_releases                  = {{ .handle_releases }}
    handle_snapshots                 = {{ .handle_snapshots }}
    suppress_pom_consistency_checks  = {{ .suppress_pom_consistency_checks }}
    reject_invalid_jars              = {{ .reject_invalid_jars }}
}

data "{{ .resource_name }}" "{{ .name }}" {
    key = {{ .resource_name }}.{{ .name }}.id
}`

func makeDataSourceRemoteGradleLikeRepoTestCase(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("%s-remote", packageType)
	resourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
	_, fqrn, name := testutil.MkNames(name, resourceName)

	params := commonJavaParams()
	params["name"] = name
	params["resource_name"] = resourceName
	params["suppress_pom_consistency_checks"] = true

	fqrn = "data." + fqrn

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, javaRepositoryBasic, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"].(string)),
					resource.TestCheckResourceAttr(fqrn, "fetch_jars_eagerly", fmt.Sprintf("%t", params["fetch_jars_eagerly"])),
					resource.TestCheckResourceAttr(fqrn, "fetch_sources_eagerly", fmt.Sprintf("%t", params["fetch_sources_eagerly"])),
					resource.TestCheckResourceAttr(fqrn, "remote_repo_checksum_policy_type", params["remote_repo_checksum_policy_type"].(string)),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%t", params["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%t", params["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%t", params["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(fqrn, "reject_invalid_jars", fmt.Sprintf("%t", params["reject_invalid_jars"])),
				),
			},
		},
	}
}

func TestAccDataSourceRemoteAllGradleLikePackageTypes(t *testing.T) {
	for _, packageType := range repository.GradleLikePackageTypes {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeDataSourceRemoteGradleLikeRepoTestCase(packageType, t))
		})
	}
}

func TestAccDataSourceRemoteMavenRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-remote", "data.artifactory_remote_maven_repository")

	params := commonJavaParams()
	params["name"] = name
	params["resource_name"] = "artifactory_remote_maven_repository"
	params["suppress_pom_consistency_checks"] = false
	params["curated"] = false

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, javaRepositoryBasic, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"].(string)),
					resource.TestCheckResourceAttr(fqrn, "fetch_jars_eagerly", fmt.Sprintf("%t", params["fetch_jars_eagerly"])),
					resource.TestCheckResourceAttr(fqrn, "fetch_sources_eagerly", fmt.Sprintf("%t", params["fetch_sources_eagerly"])),
					resource.TestCheckResourceAttr(fqrn, "remote_repo_checksum_policy_type", params["remote_repo_checksum_policy_type"].(string)),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%t", params["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%t", params["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%t", params["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(fqrn, "reject_invalid_jars", fmt.Sprintf("%t", params["reject_invalid_jars"])),
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteNugetRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nuget-remote", "data.artifactory_remote_nuget_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteNugetRepository",
		`resource "artifactory_remote_nuget_repository" "{{ .name }}" {
		    key                        = "{{ .name }}"
		    url                        = "http://tempurl.org"
		    feed_context_path          = "/foo"
		    force_nuget_authentication = true
		}

		data "artifactory_remote_nuget_repository" "{{ .name }}" {
		    key = artifactory_remote_nuget_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "nuget"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "nuget-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "feed_context_path", "/foo"),
					resource.TestCheckResourceAttr(fqrn, "download_context_path", "api/v2/package"),
					resource.TestCheckResourceAttr(fqrn, "v3_feed_url", "https://api.nuget.org/v3/index.json"),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
					resource.TestCheckResourceAttr(fqrn, "symbol_server_url", "https://symbols.nuget.org/download/symbols"),
				),
			},
		},
	})
}

func TestAccDataSourceRemotePypiRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("pypi-remote", "data.artifactory_remote_pypi_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemotePypiRepository",
		`resource "artifactory_remote_pypi_repository" "{{ .name }}" {
		    key     = "{{ .name }}"
		    url     = "http://tempurl.org"
			curated = false
		}

		data "artifactory_remote_pypi_repository" "{{ .name }}" {
		    key = artifactory_remote_pypi_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "pypi"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "pypi_registry_url", "https://pypi.org"),
					resource.TestCheckResourceAttr(fqrn, "pypi_repository_suffix", "simple"),
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteTerraformRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-remote", "data.artifactory_remote_terraform_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteTerraformRepository",
		`resource "artifactory_remote_terraform_repository" "{{ .name }}" {
		    key                  = "{{ .name }}"
		    url                  = "http://tempurl.org"
		    bypass_head_requests = true
		}

		data "artifactory_remote_terraform_repository" "{{ .name }}" {
		    key = artifactory_remote_terraform_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "terraform"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "terraform_registry_url", "https://registry.terraform.io"),
					resource.TestCheckResourceAttr(fqrn, "terraform_providers_url", "https://releases.hashicorp.com"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceRemoteVcsRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("vcs-remote", "data.artifactory_remote_vcs_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := utilsdk.ExecuteTemplate(
		"TestAccDataSourceRemoteTerraformRepository",
		`resource "artifactory_remote_vcs_repository" "{{ .name }}" {
		    key                  = "{{ .name }}"
		    url                  = "http://tempurl.org"
		    max_unique_snapshots = 2
		}

		data "artifactory_remote_vcs_repository" "{{ .name }}" {
		    key = artifactory_remote_vcs_repository.{{ .name }}.id
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "vcs"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", "2"),
				),
			},
		},
	})
}
