package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/samber/lo"
)

func NewManagedUserResource() resource.Resource {
	return &ArtifactoryManagedUserResource{
		ArtifactoryBaseUserResource: ArtifactoryBaseUserResource{
			TypeName: "artifactory_managed_user",
		},
	}
}

type ArtifactoryManagedUserResource struct {
	ArtifactoryBaseUserResource
}

func (r *ArtifactoryManagedUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	var managedUserSchemaFramework = map[string]schema.Attribute{
		"password": schema.StringAttribute{
			Required:            true,
			Sensitive:           true,
			MarkdownDescription: "Password for the user.",
		},
	}

	managedUserSchemaFramework = lo.Assign(baseUserSchemaFramework, managedUserSchemaFramework)

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory managed user resource. This can be used to create and manage Artifactory users. For example, service account where password is known and managed externally.",
		Attributes:          managedUserSchemaFramework,
		Version:             1,
	}
}
