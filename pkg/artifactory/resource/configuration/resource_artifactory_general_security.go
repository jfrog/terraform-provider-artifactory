package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"

	"gopkg.in/yaml.v3"
)

var _ resource.Resource = (*GeneralSecurityResource)(nil)

func NewGeneralSecurityResource() resource.Resource {
	return &GeneralSecurityResource{
		JFrogResource: util.JFrogResource{
			TypeName: "artifactory_general_security",
		},
	}
}

type GeneralSecurityResource struct {
	util.JFrogResource
}

type GeneralSecurityResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	EnableAnonymousAccess types.Bool   `tfsdk:"enable_anonymous_access"`
	EncryptionPolicy      types.String `tfsdk:"encryption_policy"`
}

func (r *GeneralSecurityResourceModel) toAPIModel(security *SecurityWrapperAPIModel) diag.Diagnostics {
	*security = SecurityWrapperAPIModel{
		SecurityAPIModel: SecurityAPIModel{
			AnonAccessEnabled: r.EnableAnonymousAccess.ValueBool(),
			PasswordSettings: PasswordSettingsAPIModel{
				EncryptionPolicy: r.EncryptionPolicy.ValueString(),
			},
		},
	}

	return nil
}

func (r *GeneralSecurityResourceModel) fromAPIModel(security *SecurityAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.ID = types.StringValue("security")
	r.EnableAnonymousAccess = types.BoolValue(security.AnonAccessEnabled)
	r.EncryptionPolicy = types.StringValue(security.PasswordSettings.EncryptionPolicy)

	return diags
}

type SecurityWrapperAPIModel struct {
	SecurityAPIModel `yaml:"security" json:"security"`
}

type SecurityAPIModel struct {
	AnonAccessEnabled bool                     `yaml:"anonAccessEnabled" json:"anonAccessEnabled"`
	PasswordSettings  PasswordSettingsAPIModel `yaml:"passwordSettings" json:"passwordSettings"`
}

type PasswordSettingsAPIModel struct {
	EncryptionPolicy string `yaml:"encryptionPolicy" json:"encryptionPolicy"`
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
			"encryption_policy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("SUPPORTED"),
				Validators: []validator.String{
					stringvalidator.OneOf("REQUIRED", "SUPPORTED", "UNSUPPORTED"),
				},
				MarkdownDescription: "Determines the password requirements from users identified to Artifactory from a remote client such as Maven. The options are: (1) `SUPPORTED` (default): Users can authenticate using secure encrypted passwords or clear-text passwords. (2) `REQUIRED`: Users must authenticate using secure encrypted passwords. Clear-text authentication fails. (3) `UNSUPPORTED`: Only clear-text passwords can be used for authentication. Default value is `SUPPORTED`.",
			},
		},
	}
}

func (r *GeneralSecurityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan GeneralSecurityResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var security SecurityWrapperAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(&security)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := yaml.Marshal(&security)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to marshal security settings during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData.Client)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	if plan.ID.IsUnknown() {
		plan.ID = types.StringValue("security")
	}

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

	var security SecurityAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&security).
		Get("artifactory/api/securityconfig")
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(state.fromAPIModel(&security)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	var security SecurityWrapperAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(&security)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := yaml.Marshal(&security)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to marshal security settings during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData.Client)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	if plan.ID.IsUnknown() {
		plan.ID = types.StringValue("security")
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GeneralSecurityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	resp.Diagnostics.AddWarning(
		"Security configuration cannot be deleted",
		"Artifactory does not support deletion of the security configurations.",
	)
}

// ImportState imports the resource into the Terraform state.
func (r *GeneralSecurityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
