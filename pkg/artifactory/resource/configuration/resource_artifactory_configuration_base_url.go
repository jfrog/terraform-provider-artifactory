package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

const baseUrlEndpoint = "/artifactory/api/system/configuration/baseUrl"

// NewBaseUrlResource creates a new Base URL resource.

type BaseUrlResourceModel struct {
	BaseUrl types.String `tfsdk:"base_url"`
}

func NewBaseUrlResource() resource.Resource {
	return &baseUrlResource{
		TypeName: "artifactory_configuration_base_url",
	}
}

type baseUrlResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func (r *baseUrlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *baseUrlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

// Schema defines the attributes for the resource.
func (r *baseUrlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Base URL configuration in Artifactory.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:    true,
				Description: "The Base URL for Artifactory.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.IsURLHttpOrHttps(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create sets the Base URL in Artifactory.
func (r *baseUrlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan BaseUrlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := setBaseUrl(r.ProviderData, plan.BaseUrl.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Base URL", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read retrieves the Base URL from Artifactory.
func (r *baseUrlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BaseUrlResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since there's no GET, we just return what's already in state
	resp.State.Set(ctx, state)
}

// Update modifies the Base URL in Artifactory.
func (r *baseUrlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// no Support for Updating the Base URL
	var plan BaseUrlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := setBaseUrl(r.ProviderData, plan.BaseUrl.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating Base URL", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the Base URL configuration in Artifactory.
func (r *baseUrlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// no Support for deleting the Base URL
	
}

// Helper function to set the Base URL.
func setBaseUrl(providerData util.ProviderMetadata, baseUrl string) error {

	if providerData.Client == nil {
		return fmt.Errorf("provider client is not initialized")
	}

	response, err := providerData.Client.R().
		SetHeader("Content-Type", "text/plain").
		// Set the base URL in the request body
		SetBody(baseUrl).
		Put(baseUrlEndpoint)
	if err != nil {
		return fmt.Errorf("failed to set Base URL: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("error setting Base URL: %s", response.String())
	}
	return nil
}

// Helper function to get the Base URL.
func getBaseUrl(providerData util.ProviderMetadata) (string, error) {
	response, err := providerData.Client.R().
		SetResult(map[string]string{}).
		Get(baseUrlEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get Base URL: %w", err)
	}
	if response.IsError() {
		return "", fmt.Errorf("error getting Base URL: %s", response.String())
	}
	result := response.Result().(map[string]string)
	return result["baseUrl"], nil
}

// Helper function to delete the Base URL.
func deleteBaseUrl(providerData util.ProviderMetadata) error {
	response, err := providerData.Client.R().
		Delete(baseUrlEndpoint)
	if err != nil {
		return fmt.Errorf("failed to delete Base URL: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("error deleting Base URL: %s", response.String())
	}
	return nil
}
