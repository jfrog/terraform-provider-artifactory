package security

import (
	"context"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func NewGlobalEnvironmentResource() resource.Resource {
	return &GlobalEnvironmentResource{}
}

type GlobalEnvironmentResource struct {
	ProviderData utilsdk.ProvderMetadata
}

type GlobalEnvironmentPostRequestAPIModel struct {
	Name string `json:"name"`
}

type GlobalEnvironmentPostRenameRequestAPIModel struct {
	Name string `json:"new_name"`
}

type GlobalEnvironmentsAPIModel []struct {
	Name string `json:"name"`
}

func (r *GlobalEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_global_environment"
}

// GlobalEnvironmentModel describes the Terraform resource data model to match the
// resource schema.
type GlobalEnvironmentModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *GlobalEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a global environment",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name must start with a letter and contain letters, digits and `-` character. The maximum length is 32 characters",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]*$`),
						"must start with a letter and contain letters, digits and `-` character",
					),
				},
			},
		},
	}
}

func (r *GlobalEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *GlobalEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GlobalEnvironmentModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	globalEnvPostBody := GlobalEnvironmentPostRequestAPIModel{
		Name: plan.Name.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(globalEnvPostBody).
		Post("access/api/v1/environments")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 201 Created
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	plan.Id = types.StringValue(plan.Name.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // All attributes are assigned in data
}

func (r *GlobalEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalEnvironmentModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	environments := GlobalEnvironmentsAPIModel{}

	response, err := r.ProviderData.Client.R().
		SetResult(&environments).
		Get("access/api/v1/environments")

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	var matchedEnvName *string
	for _, env := range environments {
		if env.Name == state.Id.ValueString() {
			matchedEnvName = &env.Name
		}
	}

	if matchedEnvName == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	state.Id = types.StringPointerValue(matchedEnvName)
	state.Name = types.StringPointerValue(matchedEnvName)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GlobalEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalEnvironmentModel
	var state GlobalEnvironmentModel

	// Read Terraform plan and state data into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Convert from Terraform data model into API data model
	newEnv := GlobalEnvironmentPostRenameRequestAPIModel{
		Name: plan.Name.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(newEnv).
		SetPathParam("environmentName", state.Name.ValueString()).
		Post("access/api/v1/environments/{environmentName}/rename")

	// Return error if the HTTP status code is not 200 OK
	if err != nil || response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	plan.Id = types.StringValue(newEnv.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GlobalEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalEnvironmentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		SetPathParam("environmentName", state.Id.ValueString()).
		Delete("access/api/v1/environments/{environmentName}")

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 204 OK or 404 Not Found
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusNoContent {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *GlobalEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
