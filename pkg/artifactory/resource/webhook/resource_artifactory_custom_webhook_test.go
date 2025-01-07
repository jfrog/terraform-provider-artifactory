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
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(domainValidationErrorMessageLookup[webhookType]),
			},
		},
	}
}
