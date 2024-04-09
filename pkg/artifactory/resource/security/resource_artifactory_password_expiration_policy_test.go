package security_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccPasswordExpirationPolicy_full(t *testing.T) {
	_, fqrn, policyName := testutil.MkNames("test-password-expiration-policy-full", "artifactory_password_expiration_policy")
	temp := `
	resource "artifactory_password_expiration_policy" "{{ .policyName }}" {
		name = "{{ .policyName }}"
		enabled = true
		password_max_age = {{ .passwordMaxAge }}
		notify_by_email = {{ .notifyByEmail }}
	}`

	config := util.ExecuteTemplate(policyName, temp, map[string]string{
		"policyName":     policyName,
		"passwordMaxAge": "120",
		"notifyByEmail":  "true",
	})

	updatedConfig := util.ExecuteTemplate(policyName, temp, map[string]string{
		"policyName":     policyName,
		"passwordMaxAge": "60",
		"notifyByEmail":  "false",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPasswordExpirationPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "password_max_age", "120"),
					resource.TestCheckResourceAttr(fqrn, "notify_by_email", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "password_max_age", "60"),
					resource.TestCheckResourceAttr(fqrn, "notify_by_email", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        policyName,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func testAccCheckPasswordExpirationPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		var policy security.PasswordExpirationPolicyAPIModel
		resp, err := client.R().
			SetResult(&policy).
			Get(security.PasswordExpirationPolicyEndpoint)
		if err != nil {
			return err
		}

		if resp.IsSuccess() && !policy.Enabled {
			return nil
		}

		return fmt.Errorf("password expiration policy still enabled")
	}
}
