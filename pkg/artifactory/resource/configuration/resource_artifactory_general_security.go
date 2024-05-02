package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"

	"gopkg.in/yaml.v3"
)

func NewGeneralSecurityResource() resource.Resource {
	return &GeneralSecurityResource{}
}

type GeneralSecurityResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type GeneralSecurityResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	EnableAnonymousAccess types.Bool   `tfsdk:"enable_anonymous_access"`
}

type GeneralSecurityAPIModel struct {
	GeneralSettingsAPIModel `yaml:"security" json:"security"`
}

type GeneralSettingsAPIModel struct {
	AnonAccessEnabled bool `yaml:"anonAccessEnabled" json:"anonAccessEnabled"`
}

func (r *GeneralSecurityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_general_security"
	r.TypeName = resp.TypeName
}

func (r *GeneralSecurityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"enable_anonymous_access": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func (r *GeneralSecurityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *GeneralSecurityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan GeneralSecurityResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	security := GeneralSecurityAPIModel{
		GeneralSettingsAPIModel: GeneralSettingsAPIModel{
			AnonAccessEnabled: plan.EnableAnonymousAccess.ValueBool(),
		},
	}

	content, err := yaml.Marshal(&security)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to marshal security settings during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	// we should only have one general security settings resource, using same id
	plan.ID = types.StringValue("security")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GeneralSecurityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state GeneralSecurityResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var generalSettings GeneralSettingsAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&generalSettings).
		Get("artifactory/api/securityconfig")
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	state.EnableAnonymousAccess = types.BoolValue(generalSettings.AnonAccessEnabled)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GeneralSecurityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan GeneralSecurityResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	security := GeneralSecurityAPIModel{
		GeneralSettingsAPIModel: GeneralSettingsAPIModel{
			AnonAccessEnabled: plan.EnableAnonymousAccess.ValueBool(),
		},
	}

	content, err := yaml.Marshal(&security)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to marshal security settings during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	// we should only have one general security settings resource, using same id
	plan.ID = types.StringValue("security")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GeneralSecurityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state GeneralSecurityResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content := `
security:
  anonAccessEnabled: false
`
	err := SendConfigurationPatch([]byte(content), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *GeneralSecurityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
