package webhook_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccWebhook_ArtifactLifecycle_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-artifact-lifecycle", "artifactory_artifact_lifecycle_webhook")

	params := map[string]interface{}{
		"webhookName": name,
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
				url = "https://google.com"
				secret                 = "fake-secret"
				use_secret_for_signing = true
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.1.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.use_secret_for_signing", "true"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
				),
			},
			{
				Config:                   webhookConfig,
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			}},
	})
}

func TestAccWebhook_ArtifactLifecycle(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-artifact-lifecycle", "artifactory_artifact_lifecycle_webhook")

	params := map[string]interface{}{
		"webhookName": name,
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
				url = "https://google.com"
				secret                 = "fake-secret"
				use_secret_for_signing = true
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.use_secret_for_signing", "true"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"handler.0.secret"},
			}},
	})
}
