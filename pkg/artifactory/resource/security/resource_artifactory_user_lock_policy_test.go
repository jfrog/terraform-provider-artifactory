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

func TestAccUserLockPolicy_full(t *testing.T) {
	_, fqrn, policyName := testutil.MkNames("test-user-lock-policy-full", "artifactory_user_lock_policy")
	temp := `
	resource "artifactory_user_lock_policy" "{{ .policyName }}" {
		name = "{{ .policyName }}"
		enabled = true
		login_attempts = {{ .loginAttempts }}
	}`

	config := util.ExecuteTemplate(policyName, temp, map[string]string{
		"policyName":    policyName,
		"loginAttempts": "10",
	})

	updatedConfig := util.ExecuteTemplate(policyName, temp, map[string]string{
		"policyName":    policyName,
		"loginAttempts": "20",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserLockPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "login_attempts", "10"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "login_attempts", "20"),
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

func testAccCheckUserLockPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		var policy security.UserLockPolicyAPIModel
		resp, err := client.R().
			SetResult(&policy).
			Get(security.UserLockPolicyEndpoint)
		if err != nil {
			return err
		}

		if resp.IsSuccess() && !policy.Enabled {
			return nil
		}

		return fmt.Errorf("user lock policy still enabled")
	}
}
