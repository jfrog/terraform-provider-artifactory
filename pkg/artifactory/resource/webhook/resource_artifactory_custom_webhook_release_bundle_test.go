package webhook_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccCustomWebhook_Distribution(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if !strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is not a cloud instance. It also needs to have distribution enabled.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("test-distribution-custom-webhook", "artifactory_distribution_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
		"payload":     `{ \"event_type\": \"distribution/{{ .type }}\", \"client_payload\": {{ json .data}} }`,
	}
	webhookConfig := util.ExecuteTemplate("TestAccCustomWebhookDistribution", `
		resource "artifactory_distribution_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			enabled     = true
			event_types = ["distribute_started", "distribute_completed", "delete_started", "delete_completed"]

			criteria {
				any_release_bundle              = true
				registered_release_bundle_names = []
				include_patterns                = ["**/*"]
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
				payload = "{{ .payload }}"
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
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "4"),
					resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", "distribute_started"),
					resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", "distribute_completed"),
					resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", "delete_started"),
					resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", "delete_completed"),
					resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.any_release_bundle", "true"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.registered_release_bundle_names.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "handler.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.method", "POST"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret1", "value1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret2", "value2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-1", "value-1"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-2", "value-2"),
					resource.TestCheckResourceAttr(fqrn, "handler.0.payload", "{ \"event_type\": \"distribution/{{ .type }}\", \"client_payload\": {{ json .data}} }"),
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
