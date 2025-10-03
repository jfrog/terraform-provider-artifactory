package remote

import (
	"context"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

func NewTerraformRemoteRepositoryResource() resource.Resource {
	return &remoteTerraformResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.TerraformPackageType,
			repository.PackageNameLookup[repository.TerraformPackageType],
			reflect.TypeFor[remoteTerraformResourceModel](),
			reflect.TypeFor[RemoteTerraformAPIModel](),
		),
	}
}

type remoteTerraformResource struct {
	remoteResource
}

type remoteTerraformResourceModel struct {
	RemoteResourceModel
	TerraformRegistryURL  types.String `tfsdk:"terraform_registry_url"`
	TerraformProvidersURL types.String `tfsdk:"terraform_providers_url"`
}

func (r *remoteTerraformResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteTerraformResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteTerraformResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteTerraformResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteTerraformResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteTerraformResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteTerraformResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteTerraformResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteTerraformAPIModel{
		RemoteAPIModel:        remoteAPIModel,
		TerraformRegistryURL:  r.TerraformRegistryURL.ValueString(),
		TerraformProvidersURL: r.TerraformProvidersURL.ValueString(),
	}, diags
}

func (r *remoteTerraformResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteTerraformAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.TerraformRegistryURL = types.StringValue(model.TerraformRegistryURL)
	r.TerraformProvidersURL = types.StringValue(model.TerraformProvidersURL)
	return diags
}

type RemoteTerraformAPIModel struct {
	RemoteAPIModel
	TerraformRegistryURL  string `json:"terraformRegistryUrl"`
	TerraformProvidersURL string `json:"terraformProvidersUrl"`
}

const WizRegistryBypassSupportedVersion = "7.122.0"

// Custom validator for bypass_head_requests to enforce true for specific terraform registries
type terraformBypassHeadRequestsValidator struct {
	providerData *util.ProviderMetadata
}

func (v terraformBypassHeadRequestsValidator) Description(ctx context.Context) string {
	return "Validates that bypass_head_requests is true for terraform registries that require it"
}

func (v terraformBypassHeadRequestsValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `bypass_head_requests` is `true` for terraform registries that require it (registry.terraform.io, registry.opentofu.org, tf.app.wiz.io for Artifactory 7.122.0+)"
}

func (v terraformBypassHeadRequestsValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Get the terraform_registry_url from the same resource
	var registryURL types.String
	diags := req.Config.GetAttribute(ctx, path.Root("terraform_registry_url"), &registryURL)
	if diags.HasError() {
		return
	}

	registryURLValue := strings.TrimSuffix(registryURL.ValueString(), "/")

	// Always enforce bypass_head_requests = true for these registries
	alwaysRequiredRegistries := []string{
		"https://registry.terraform.io",
		"https://registry.opentofu.org",
	}

	for _, requiredRegistry := range alwaysRequiredRegistries {
		if registryURLValue == requiredRegistry {
			if !req.ConfigValue.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid bypass_head_requests value",
					"For terraform registries (registry.terraform.io, registry.opentofu.org), bypass_head_requests must be set to true. Artifactory automatically enforces this setting for these registries.",
				)
			}
			return
		}
	}

	// Check tf.app.wiz.io registry with version requirement
	if registryURLValue == "https://tf.app.wiz.io" {
		// Check if provider data is available
		if v.providerData == nil {
			// If no provider data available, skip validation (shouldn't happen in normal operation)
			return
		}

		// Check if Artifactory version supports wiz registry bypass
		isSupported, err := util.CheckVersion(v.providerData.ArtifactoryVersion, WizRegistryBypassSupportedVersion)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Failed to check Artifactory version",
				"Unable to validate bypass_head_requests requirement for tf.app.wiz.io: "+err.Error(),
			)
			return
		}

		if isSupported {
			// For Artifactory 7.122.0+, enforce bypass_head_requests = true for wiz registry
			if !req.ConfigValue.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid bypass_head_requests value",
					"For tf.app.wiz.io registry with Artifactory 7.122.0 or later, bypass_head_requests must be set to true. Artifactory automatically enforces this setting for this registry.",
				)
			}
		}
		// For versions before 7.122.0, no validation needed for wiz registry
		return
	}
}

func (r *remoteTerraformResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteTerraformAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"bypass_head_requests": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Validators: []validator.Bool{
					terraformBypassHeadRequestsValidator{providerData: r.ProviderData},
				},
				MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. " +
					"In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the " +
					"artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request. " +
					"**Note**: For terraform registries (registry.terraform.io, registry.opentofu.org), this must be set to `true` as Artifactory automatically enforces this setting. " +
					"For tf.app.wiz.io, this is required only for Artifactory 7.122.0 and later. Defaults to `false`.",
			},
			"terraform_registry_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://registry.terraform.io"),
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "The base URL of the registry API. When using Smart Remote Repositories, set the URL to" +
					" <base_Artifactory_URL>/api/terraform/repokey. Default value in UI is https://registry.terraform.io",
			},
			"terraform_providers_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://releases.hashicorp.com"),
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "The base URL of the Provider's storage API. When using Smart remote repositories, set " +
					"the URL to <base_Artifactory_URL>/api/terraform/repokey/providers. Default value in UI is https://releases.hashicorp.com",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteTerraformAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
