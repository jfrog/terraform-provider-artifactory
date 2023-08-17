package user

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"golang.org/x/net/context"
)

func NewAnonymousUserResource() resource.Resource {

	return &ArtifactoryAnonymousUserResource{}
}

type ArtifactoryAnonymousUserResource struct {
	client utilsdk.ProvderMetadata
}

// ArtifactoryAnonymousUserResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryAnonymousUserResourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// ArtifactoryAnonymousUserResourceAPIModel describes the API data model.
type ArtifactoryAnonymousUserResourceAPIModel struct {
	Name string `json:"name"`
}

func (r *ArtifactoryAnonymousUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_anonymous_user"
}

func (r *ArtifactoryAnonymousUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory anonymous user resource. This can be used to import Artifactory 'anonymous' uer for some use cases where this is useful.\n\nThis resource is not intended for managing the 'anonymous' user in Artifactory. Use the `resource_artifactory_user` resource instead.\n\n!> Anonymous user cannot be created from scratch, nor updated/deleted once imported into Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Username for anonymous user. This is only for ensuring resource schema is valid for Terraform. This is not meant to be set or updated in the HCL.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ArtifactoryAnonymousUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *ArtifactoryAnonymousUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Unable to Create Resource",
		"Anonymous Artifactory user cannot be created. Use `terraform import` instead.",
	)
	return
}

func (r *ArtifactoryAnonymousUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ArtifactoryAnonymousUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	user := &ArtifactoryAnonymousUserResourceAPIModel{}

	response, err := r.client.Client.R().SetResult(user).Get(UsersEndpointPath + data.Id.ValueString())

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	data.toState(user)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryAnonymousUserResourceModel) toState(user *ArtifactoryAnonymousUserResourceAPIModel) {
	r.Id = types.StringValue(user.Name)
	r.Name = types.StringValue(user.Name)
}

func (r *ArtifactoryAnonymousUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unable to Update Resource",
		"Anonymous Artifactory user cannot be updated. Use `terraform import` instead.",
	)
	return
}

func (r *ArtifactoryAnonymousUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Unable to Delete Resource",
		"Anonymous Artifactory user cannot be deleted. Use `terraform state rm` instead.",
	)
	return
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactoryAnonymousUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
