package webhook

import "github.com/hashicorp/terraform-plugin-framework/resource"

var _ resource.Resource = &ArtifactWebhookResource{}

func NewArtifactWebhookResource() resource.Resource {
	return &ArtifactWebhookResource{
		TypeName: "artifactory_artifact_webhook",
	}
}
