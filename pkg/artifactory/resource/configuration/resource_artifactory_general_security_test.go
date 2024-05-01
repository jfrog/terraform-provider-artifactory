package configuration_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

const generalSecurityTemplateFull = `
resource "artifactory_general_security" "security" {
	enable_anonymous_access = true
}`

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
				ConfigPlanChecks: testutil.ConfigPlanChecks,
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   generalSecurityTemplateFull,
				PlanOnly:                 true,
				ConfigPlanChecks:         testutil.ConfigPlanChecks,
			},
		},
	})
}

func TestAccGeneralSecurity_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	fqrn := "artifactory_general_security.security"

	temp := `
	resource "artifactory_general_security" "security" {
		enable_anonymous_access = {{ .enableAnonymousAccess }}
	}`

	config := util.ExecuteTemplate(
		"TestAccGeneralSecurity_full",
		temp,
		map[string]interface{}{
			"enableAnonymousAccess": true,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		"TestAccGeneralSecurity_full",
		temp,
		map[string]interface{}{
			"enableAnonymousAccess": false,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccGeneralSecurityDestroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "false"),
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

func testAccGeneralSecurityDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		var generalSettings configuration.GeneralSettingsAPIModel
		resp, err := client.R().SetResult(&generalSettings).Get("artifactory/api/securityconfig")
		if err != nil || resp.IsError() {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
		}
		if generalSettings.AnonAccessEnabled != false {
			return fmt.Errorf("error: general security setting to allow anonymous access is still enabled")
		}

		return nil
	}
}
