package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewDockerRemoteRepositoryResource() resource.Resource {
	return &remoteDockerResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.DockerPackageType,
			repository.PackageNameLookup[repository.DockerPackageType],
			reflect.TypeFor[remoteDockerResourceModel](),
			reflect.TypeFor[RemoteDockerAPIModel](),
		),
	}
}

type remoteDockerResource struct {
	remoteResource
}

type remoteDockerResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	ExternalDependenciesEnabled  types.Bool   `tfsdk:"external_dependencies_enabled"`
	ExternalDependenciesPatterns types.List   `tfsdk:"external_dependencies_patterns"`
	EnableTokenAuthentication    types.Bool   `tfsdk:"enable_token_authentication"`
	BlockPushingSchema1          types.Bool   `tfsdk:"block_pushing_schema1"`
	ProjectId                    types.String `tfsdk:"project_id"`
}

func (r *remoteDockerResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteDockerResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteDockerResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteDockerResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteDockerResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteDockerResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteDockerResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteDockerResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	var externalDependenciesPatterns []string
	d = r.ExternalDependenciesPatterns.ElementsAs(ctx, &externalDependenciesPatterns, false)
	if d != nil {
		diags.Append(d...)
	}

	var apiModel = RemoteDockerAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
		ExternalDependenciesEnabled: r.ExternalDependenciesEnabled.ValueBool(),
		EnableTokenAuthentication:   r.EnableTokenAuthentication.ValueBool(),
		BlockPushingSchema1:         r.BlockPushingSchema1.ValueBool(),
		ProjectId:                   r.ProjectId.ValueString(),
	}
	if r.ExternalDependenciesEnabled.ValueBool() == true {
		apiModel.ExternalDependenciesPatterns = externalDependenciesPatterns
	}
	return apiModel, diags
}

func (r *remoteDockerResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteDockerAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.ExternalDependenciesEnabled = types.BoolValue(model.ExternalDependenciesEnabled)

	if r.ExternalDependenciesEnabled.ValueBool() == true {
		externalDependenciesPatterns, d := types.ListValueFrom(ctx, types.StringType, model.ExternalDependenciesPatterns)
		if d != nil {
			diags.Append(d...)
		}
		r.ExternalDependenciesPatterns = externalDependenciesPatterns
	}

	r.EnableTokenAuthentication = types.BoolValue(model.EnableTokenAuthentication)
	r.BlockPushingSchema1 = types.BoolValue(model.BlockPushingSchema1)
	r.ProjectId = types.StringValue(model.ProjectId)

	return diags
}

type RemoteDockerAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `json:"blockPushingSchema1"`
	ProjectId                    string   `json:"dockerProjectId"`
}

func (r *remoteDockerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteDockerAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"external_dependencies_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Also known as 'Foreign Layers Caching' on the UI, default is `false`.",
			},
			"enable_token_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable token (Bearer) based authentication.",
			},
			"block_pushing_schema1": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, Artifactory will block the pulling of Docker images with manifest v2 schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1 that exist in the cache.",
			},
			"external_dependencies_patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.AlsoRequires(path.MatchRoot("external_dependencies_enabled")),
					listvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
					"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
					"This attribute must be set together with `external_dependencies_enabled = true`",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id will default to their default project as you have set up in your account.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteDockerAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}

func (r remoteDockerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data remoteDockerResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If external_dependencies_patterns is not configured, return without warning.
	if data.ExternalDependenciesPatterns.IsNull() || data.ExternalDependenciesPatterns.IsUnknown() {
		return
	}

	// If external_dependencies_enabled is not null, return without warning.
	if !data.ExternalDependenciesPatterns.IsNull() && !data.ExternalDependenciesEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("external_dependencies_patterns"),
			"Invalid Attribute Configuration",
			"external_dependencies_enabled must set to 'true' when external_dependencies_patterns is configured.",
		)
	}
}
