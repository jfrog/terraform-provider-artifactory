package configuration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccGeneralSecurity_UpgradeFromSDKv2(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	fqrn := "artifactory_general_security.security"
	config := `
	resource "artifactory_general_security" "security" {
		enable_anonymous_access = true
	}`

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "10.7.4",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "true"),
				),
				ConfigPlanChecks: testutil.ConfigPlanChecks(""),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.7.1",
						Source:            "jfrog/artifactory",
					},
				},
				Config:           config,
				PlanOnly:         true,
				ConfigPlanChecks: testutil.ConfigPlanChecks(""),
			},
		},
	})
}

func TestAccGeneralSecurity_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, resourceName := testutil.MkNames("test-general-security", "artifactory_general_security")

	temp := `
	resource "artifactory_general_security" "{{ .name }}" {
		enable_anonymous_access = {{ .enableAnonymousAccess }}
		encryption_policy = "{{ .encryption_policy }}"
	}`

	config := util.ExecuteTemplate(
		"TestAccGeneralSecurity_full",
		temp,
		map[string]interface{}{
			"name":                  resourceName,
			"enableAnonymousAccess": false,
			"encryption_policy":     "UNSUPPORTED",
		},
	)

	updatedConfig := util.ExecuteTemplate(
		"TestAccGeneralSecurity_full",
		temp,
		map[string]interface{}{
			"name":                  resourceName,
			"enableAnonymousAccess": true,
			"encryption_policy":     "SUPPORTED",
		},
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "encryption_policy", "UNSUPPORTED"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "true"),
					resource.TestCheckResourceAttr(fqrn, "encryption_policy", "SUPPORTED"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
