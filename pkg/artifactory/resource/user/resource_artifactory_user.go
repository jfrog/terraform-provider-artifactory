package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/samber/lo"
)

func NewUserResource() resource.Resource {
	return &ArtifactoryUserResource{
		ArtifactoryBaseUserResource: ArtifactoryBaseUserResource{
			TypeName: "artifactory_user",
		},
	}
}

type ArtifactoryUserResource struct {
	ArtifactoryBaseUserResource
}

func (r *ArtifactoryUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	var userSchemaFramework = map[string]schema.Attribute{
		"password": schema.StringAttribute{
			MarkdownDescription: "(Optional, Sensitive) Password for the user. When omitted, a random password is generated using the following password policy: " +
				"12 characters with 1 digit, 1 symbol, with upper and lower case letters",
			Optional:  true,
			Sensitive: true,
			Computed:  true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	userSchemaFramework = lo.Assign(baseUserSchemaFramework, userSchemaFramework)

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory user resource. This can be used to create and manage Artifactory users. The password is a required field by the [Artifactory API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplaceUser), but we made it optional in this resource to accommodate the scenario where the password is not needed and will be reset by the actual user later. When the optional attribute `password` is omitted, a random password is generated according to current Artifactory password policy.",
		Attributes:          userSchemaFramework,
	}
}
