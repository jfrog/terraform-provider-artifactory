package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccWebhook_ReleaseBundleV2_NilHandling(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_release_bundle_v2_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhook_ReleaseBundleV2_NilHandling", `
		resource "artifactory_release_bundle_v2_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["release_bundle_v2_started"]
			criteria {
				any_release_bundle = true
				selected_release_bundles = []
			}
			handler {
				url = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "criteria.0.any_release_bundle", "true"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.selected_release_bundles.#", "0"),
				),
			},
		},
	})
}
