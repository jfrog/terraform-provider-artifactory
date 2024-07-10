package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccWebhook_ArtifactLifecycle(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-artifact-lifecycle", "artifactory_artifact_lifecycle_webhook")

	params := map[string]interface{}{
		"webhookName":         name,
		"useSecretForSigning": testutil.RandBool(),
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhook_User", `
		resource "artifactory_artifact_lifecycle_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = [
				"archive",
				"restore",
			]
			handler {
				url = "https://tempurl.org"
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
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://tempurl.org"),
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
