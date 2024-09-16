package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccWebhook_ReleaseBundleV2Promotion(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-release-bundle-v2-promotion", "artifactory_release_bundle_v2_promotion_webhook")

	params := map[string]interface{}{
		"webhookName":         name,
		"useSecretForSigning": testutil.RandBool(),
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhook_User", `
		resource "artifactory_release_bundle_v2_promotion_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = [
				"release_bundle_v2_promotion_completed",
				"release_bundle_v2_promotion_failed",
				"release_bundle_v2_promotion_started",
			]
			criteria {
				selected_environments = [
					"PROD",
					"DEV",
				]
			}
			handler {
				url = "https://google.com"
				secret                 = "fake-secret"
				use_secret_for_signing = {{ .useSecretForSigning }}
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, "", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "3"),
					resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.selected_environments.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "PROD"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.use_secret_for_signing", fmt.Sprintf("%t", params["useSecretForSigning"])),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"handler.0.secret"},
			}},
	})
}
