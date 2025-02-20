package local_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLocalAlpineRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("alpine-local-test-repo", "artifactory_local_alpine_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("keypair", temp, testData) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	updatedTemp := `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			project_environments = []
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`

	updatedConfig := util.ExecuteTemplate("keypair", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.AlpinePackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccLocalAlpineRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("alpine-local-test-repo", "artifactory_local_alpine_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	config := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				// ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccLocalAlpineRepository_UpgradeFromSDKv2_EmptyProjectEnvironments(t *testing.T) {
	_, fqrn, name := testutil.MkNames("alpine-local-test-repo", "artifactory_local_alpine_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	config := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			project_environments = []
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.6.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
				// ConfigPlanChecks: resource.ConfigPlanChecks{
				// 	PreApply: []plancheck.PlanCheck{
				// 		plancheck.ExpectEmptyPlan(),
				// 	},
				// },
			},
		},
	})
}
