package security

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	VaultConfigurationsEndpoint = "access/api/v1/vault/configs"
	VaultConfigurationEndpoint  = "access/api/v1/vault/configs/{name}"
)

var _ resource.Resource = &VaultConfigurationResource{}

func NewVaultConfigurationResource() resource.Resource {
	return &VaultConfigurationResource{
		TypeName: "artifactory_vault_configuration",
	}
}

type VaultConfigurationResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type VaultConfigurationResourceModel struct {
	Name   types.String `tfsdk:"name"`
	Config types.Object `tfsdk:"config"`
}

func (m VaultConfigurationResourceModel) toAPIModel(_ context.Context, apiModel *VaultConfigurationAPIModel) (diags diag.Diagnostics) {
	configAttrs := m.Config.Attributes()
	authAttrs := configAttrs["auth"].(types.Object).Attributes()

	auth := VaultConfigurationConfigAuthAPIModel{
		Type: authAttrs["type"].(types.String).ValueString(),
	}

	if auth.Type == "Certificate" {
		auth.Certificate = authAttrs["certificate"].(types.String).ValueString()
		auth.CertificateKey = authAttrs["certificate_key"].(types.String).ValueString()
	}

	if auth.Type == "AppRole" {
		auth.RoleID = authAttrs["role_id"].(types.String).ValueString()
		auth.SecretID = authAttrs["secret_id"].(types.String).ValueString()
	}

	*apiModel = VaultConfigurationAPIModel{
		Type: "HashicorpVault",
		Config: VaultConfigurationConfigAPIModel{
			URL:  configAttrs["url"].(types.String).ValueString(),
			Auth: auth,
			Mounts: lo.Map(
				configAttrs["mounts"].(types.Set).Elements(),
				func(elem attr.Value, _ int) VaultConfigurationConfigMountAPIModel {
					attrs := elem.(types.Object).Attributes()

					return VaultConfigurationConfigMountAPIModel{
						Path: attrs["path"].(types.String).ValueString(),
						Type: attrs["type"].(types.String).ValueString(),
					}
				},
			),
		},
	}

	return
}

var vaultConfigurationConfigAttributeModel = map[string]attr.Type{
	"url":    types.StringType,
	"auth":   types.ObjectType{AttrTypes: configAuthResourceModelAttributeTypes},
	"mounts": types.SetType{ElemType: configMountSetResourceModelAttributeTypes},
}

var configAuthResourceModelAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"type":            types.StringType,
	"certificate":     types.StringType,
	"certificate_key": types.StringType,
	"role_id":         types.StringType,
	"secret_id":       types.StringType,
}

var configMountResourceModelAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"path": types.StringType,
	"type": types.StringType,
}

var configMountSetResourceModelAttributeTypes types.ObjectType = types.ObjectType{
	AttrTypes: configMountResourceModelAttributeTypes,
}

func (m *VaultConfigurationResourceModel) fromAPIModel(ctx context.Context, apiModel VaultConfigurationAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.Name = types.StringValue(apiModel.Key)

	certificate := types.StringNull()
	if len(apiModel.Config.Auth.Certificate) > 0 {
		certificate = types.StringValue(apiModel.Config.Auth.Certificate)
	}

	certificateKey := types.StringNull()
	if len(apiModel.Config.Auth.CertificateKey) > 0 {
		certificateKey = types.StringValue(apiModel.Config.Auth.CertificateKey)
	}

	roleID := types.StringNull()
	secretID := types.StringNull()
	configAttrs := m.Config.Attributes()
	if v, ok := configAttrs["auth"]; ok {
		authAttrs := v.(types.Object).Attributes()
		if r, ok := authAttrs["role_id"]; ok {
			roleID = r.(types.String)
		}
		if s, ok := authAttrs["secret_id"]; ok {
			secretID = s.(types.String)
		}
	}

	auth, ds := types.ObjectValue(
		configAuthResourceModelAttributeTypes,
		map[string]attr.Value{
			"type":            types.StringValue(apiModel.Config.Auth.Type),
			"certificate":     certificate,
			"certificate_key": certificateKey,
			"role_id":         roleID,   // use resource value as API returns hashed value
			"secret_id":       secretID, // use resource value as API returns hashed value
		},
	)
	if ds.HasError() {
		diags.Append(ds...)
	}

	mounts := lo.Map(
		apiModel.Config.Mounts,
		func(property VaultConfigurationConfigMountAPIModel, _ int) attr.Value {
			mount, ds := types.ObjectValue(
				configMountResourceModelAttributeTypes,
				map[string]attr.Value{
					"path": types.StringValue(property.Path),
					"type": types.StringValue(property.Type),
				},
			)

			if ds != nil {
				diags.Append(ds...)
			}

			return mount
		},
	)
	mountsSet, ds := types.SetValueFrom(
		ctx,
		configMountSetResourceModelAttributeTypes,
		mounts,
	)
	if ds != nil {
		diags.Append(ds...)
	}

	config, ds := types.ObjectValue(
		vaultConfigurationConfigAttributeModel,
		map[string]attr.Value{
			"url":    types.StringValue(apiModel.Config.URL),
			"auth":   auth,
			"mounts": mountsSet,
		},
	)
	if ds.HasError() {
		diags.Append(ds...)
	}

	m.Config = config

	return diags
}

type VaultConfigurationAPIModel struct {
	Key    string                           `json:"key,omitempty"`
	Type   string                           `json:"type"`
	Config VaultConfigurationConfigAPIModel `json:"config"`
}

type VaultConfigurationConfigAPIModel struct {
	URL    string                                  `json:"url"`
	Auth   VaultConfigurationConfigAuthAPIModel    `json:"auth"`
	Mounts []VaultConfigurationConfigMountAPIModel `json:"mounts"`
}

type VaultConfigurationConfigAuthAPIModel struct {
	Type           string `json:"type"`
	Certificate    string `json:"certificate,omitempty"`
	CertificateKey string `json:"certificateKey,omitempty"`
	RoleID         string `json:"roleId,omitempty"`
	SecretID       string `json:"secretId,omitempty"`
}

type VaultConfigurationConfigMountAPIModel struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func (r *VaultConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *VaultConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the Vault configuration",
			},
			"config": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							validatorfw_string.IsURLHttpOrHttps(),
						},
						Description: "The base URL of the Vault server.",
					},
					"auth": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("Certificate", "AppRole", "Agent"),
								},
								MarkdownDescription: "The authentication method used. The supported methods are `Certificate`, `AppRole`, and `Agent`. For more information, see [Hashicorp Vault Docs](https://developer.hashicorp.com/vault/docs/auth).",
							},
							"certificate": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("certificate_key")),
								},
								MarkdownDescription: "Client certificate (in PEM format) for `Certificate` type.",
							},
							"certificate_key": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("certificate")),
								},
								MarkdownDescription: "Private key (in PEM format) for `Certificate` type.",
							},
							"role_id": schema.StringAttribute{
								Optional:  true,
								Sensitive: true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_id")),
								},
								MarkdownDescription: "Role ID for `AppRole` type",
							},
							"secret_id": schema.StringAttribute{
								Optional:  true,
								Sensitive: true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("role_id")),
								},
								MarkdownDescription: "Secret ID for `AppRole` type",
							},
						},
						Required: true,
					},
					"mounts": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Required:    true,
									Description: "Vault secret engine path",
								},
								"type": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("KV1", "KV2"),
									},
									MarkdownDescription: "Vault supports several secret engines, each one has different capabilities. The supported secret engine types are: `KV1` and `KV2`.",
								},
							},
						},
						Required: true,
					},
				},
				Required: true,
			},
		},
		MarkdownDescription: "This resource enables you to configure an external vault connector to use as a centralized secret management tool for the keys used to sign packages. For more information, see [JFrog documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/vault).\nThis feature is supported with Enterprise X and Enterprise+ licenses.",
	}
}

func (r VaultConfigurationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data VaultConfigurationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configAttrs := data.Config.Attributes()
	authAttrs := configAttrs["auth"].(types.Object).Attributes()
	authType := authAttrs["type"].(types.String)

	switch authType.ValueString() {
	case "Certificate":
		if v, ok := authAttrs["certificate"]; !ok || v.IsNull() || v.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("config").AtName("auth").AtName("certificate"),
				"Missing Attribute Configuration",
				"Expected 'certificate' to be configured when auth type set to 'Certificate'.",
			)
		}
		if v, ok := authAttrs["certificate_key"]; !ok || v.IsNull() || v.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("config").AtName("auth").AtName("certificate_key"),
				"Missing Attribute Configuration",
				"Expected 'certificate_key' to be configured when auth type set to 'Certificate'.",
			)
		}

	case "AppRole":
		if v, ok := authAttrs["role_id"]; !ok || v.IsNull() || v.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("config").AtName("auth").AtName("role_id"),
				"Missing Attribute Configuration",
				"Expected 'role_id' to be configured when auth type set to 'AppRole'.",
			)
		}
		if v, ok := authAttrs["secret_id"]; !ok || v.IsNull() || v.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("config").AtName("auth").AtName("secret_id"),
				"Missing Attribute Configuration",
				"Expected 'secret_id' to be configured when auth type set to 'AppRole'.",
			)
		}
	}
}

func (r *VaultConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *VaultConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan VaultConfigurationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vaultConfig VaultConfigurationAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &vaultConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("name", plan.Name.ValueString()).
		SetBody(vaultConfig).
		Put(VaultConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}
	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VaultConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state VaultConfigurationResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vaultConfigs []VaultConfigurationAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&vaultConfigs).
		Get(VaultConfigurationsEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	matchedVaultConfig, ok := lo.Find(
		vaultConfigs,
		func(config VaultConfigurationAPIModel) bool {
			return config.Key == state.Name.ValueString()
		},
	)

	if !ok {
		utilfw.UnableToRefreshResourceError(resp, fmt.Sprintf("Unable to find Vault configuration %s", state.Name.ValueString()))
		return
	}

	resp.Diagnostics.Append(state.fromAPIModel(ctx, matchedVaultConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VaultConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan VaultConfigurationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vaultConfig VaultConfigurationAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &vaultConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("name", plan.Name.ValueString()).
		SetBody(vaultConfig).
		Put(VaultConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}
	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VaultConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state VaultConfigurationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		SetPathParam("name", state.Name.ValueString()).
		Delete(VaultConfigurationEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *VaultConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
