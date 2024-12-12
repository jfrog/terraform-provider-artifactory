package webhook_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccCustomWebhook_ArtifactLifecycle_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-artifact-lifecycle-webhook", "artifactory_artifact_lifecycle_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
	}
	config := util.ExecuteTemplate("TestAccCustomWebhookArtifactLifecycle", `
		resource "artifactory_artifact_lifecycle_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["archive", "restore"]
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
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
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "2"),
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
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCustomWebhook_ArtifactLifecycle(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-artifact-lifecycle-webhook", "artifactory_artifact_lifecycle_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccCustomWebhookArtifactLifecycle", `
		resource "artifactory_artifact_lifecycle_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["archive", "restore"]
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "2"),
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
