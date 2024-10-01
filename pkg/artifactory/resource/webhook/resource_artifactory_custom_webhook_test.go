package webhook_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccCustomWebhook_CriteriaValidation(t *testing.T) {
	for _, webhookType := range []string{webhook.ArtifactDomain, webhook.ArtifactPropertyDomain, webhook.ArtifactoryReleaseBundleDomain, webhook.BuildDomain, webhook.DestinationDomain, webhook.DistributionDomain, webhook.DockerDomain, webhook.ReleaseBundleDomain, webhook.ReleaseBundleV2Domain} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(customWebhookCriteriaValidationTestCase(webhookType, t))
		})
	}
}

func customWebhookCriteriaValidationTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_custom_webhook.%s", webhookType, name)

	var template string
	switch webhookType {
	case "artifact", "artifact_property", "docker":
		template = repoTemplate
	case "build":
		template = buildTemplate
	case "release_bundle", "distribution", "artifactory_release_bundle", "destination":
		template = releaseBundleTemplate
	case "release_bundle_v2":
		template = releaseBundleV2Template
	}

	params := map[string]interface{}{
		"webhookType": webhookType,
		"webhookName": name,
		"eventTypes":  webhook.DomainEventTypesSupported[webhookType],
	}
	webhookConfig := util.ExecuteTemplate("TestAccCustomWebhookCriteriaValidation", template, params)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(domainValidationErrorMessageLookup[webhookType]),
			},
		},
	}
}

func TestAccCustomWebhook_User(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-user-webhook", "artifactory_user_custom_webhook")

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccCustomWebhookUser", `
		resource "artifactory_user_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["locked"]
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, "", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "event_types.#", "1"),
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
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"handler.0.secrets"},
			},
		},
	})
}
