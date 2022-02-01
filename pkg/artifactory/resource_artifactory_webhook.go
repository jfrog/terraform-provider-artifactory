package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var webhookTypesSupported = []string{
	"artifact",
	"artifact_property",
	"build",
	"release_bundle",
	"distribution",
	"artifactory_release_bundle",
}

func resourceArtifactoryWebhook(webhookType string) *schema.Resource {
	return mkResourceSchema(legacyLocalSchema, inSchema(legacyLocalSchema), unmarshalLocalRepository, func() interface{} {
		return &MessyLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				Rclass: "local",
			},
		}
	})
}
