package webhook_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccCustomWebhook_ReleaseBundleV2Promotion_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-release-bundle-v2-promotion-webhook", "artifactory_release_bundle_v2_promotion_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
	}
	config := util.ExecuteTemplate("TestAccCustomWebhookReleaseBundleV2Promotion", `
		resource "artifactory_release_bundle_v2_promotion_custom_webhook" "{{ .webhookName }}" {
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
				secrets = {
					secret1 = "value1"
					secret2 = "value2"
				}
				http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.1.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "3"),
					resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.selected_environments.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "PROD"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret1", "value1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret2", "value2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-2", "value-2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"),
				),
			},
			{
				Config:                   config,
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCustomWebhook_ReleaseBundleV2Promotion(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-release-bundle-v2-promotion-webhook", "artifactory_release_bundle_v2_promotion_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccCustomWebhookReleaseBundleV2Promotion", `
		resource "artifactory_release_bundle_v2_promotion_custom_webhook" "{{ .webhookName }}" {
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
				method = "POST"
				secrets = {
					secret1 = "value1"
					secret2 = "value2"
				}
				http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "3"),
					resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.selected_environments.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "PROD"),
					resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.selected_environments.*", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.method", "POST"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret1", "value1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret2", "value2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-2", "value-2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"handler.0.secrets"},
			},
		},
	})
}
