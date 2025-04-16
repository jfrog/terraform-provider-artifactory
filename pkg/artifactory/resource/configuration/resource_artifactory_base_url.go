package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
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

// Schema defines the attributes for the resource.
func (r *baseUrlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Base URL configuration in Artifactory.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:    true,
				Description: "The Base URL for Artifactory.",
			},
		},
	}
}

// Create sets the Base URL in Artifactory.
func (r *baseUrlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
func (r *baseUrlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(util.ProviderMetadata)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected util.ProviderMetadata but got %T", req.ProviderData),
		)
		return
	}

	r.ProviderData = providerData
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
	err := deleteBaseUrl(r.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Base URL", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Helper function to set the Base URL.
func setBaseUrl(providerMeta util.ProviderMetadata, baseUrl string) error {
	client := providerMeta.Client

	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(baseUrl).
		Put("/artifactory/api/system/configuration/baseUrl")

	if err != nil {
		return fmt.Errorf("error setting Base URL: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error setting Base URL: %s", resp.String())
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
