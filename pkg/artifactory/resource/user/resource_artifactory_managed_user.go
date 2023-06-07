package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/maps"
)

func NewManagedUserResource() resource.Resource {
	return &ArtifactoryManagedUserResource{}
}

type ArtifactoryManagedUserResource struct {
	ArtifactoryBaseUserResource
}

func (r *ArtifactoryManagedUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_managed_user"
}

func (r *ArtifactoryManagedUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	var managedUserSchemaFramework = map[string]schema.Attribute{
		"password": schema.StringAttribute{
			MarkdownDescription: "Password for the user.",
			Required:            true,
			Sensitive:           true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(12)},
		},
	}

	maps.Copy(managedUserSchemaFramework, baseUserSchemaFramework)

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory managed user resource. This can be used to create and manage Artifactory users. For example, service account where password is known and managed externally.",
		Attributes:          managedUserSchemaFramework,
	}
}
