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

func NewHelmOCIRemoteRepositoryResource() resource.Resource {
	return &remoteHelmOCIResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.HelmOCIPackageType,
			repository.PackageNameLookup[repository.HelmOCIPackageType],
			reflect.TypeFor[remoteHelmOCIResourceModel](),
			reflect.TypeFor[RemoteHelmOCIAPIModel](),
		),
	}
}

type remoteHelmOCIResource struct {
	remoteResource
}

type remoteHelmOCIResourceModel struct {
	RemoteResourceModel
	ExternalDependenciesEnabled  types.Bool   `tfsdk:"external_dependencies_enabled"`
	ExternalDependenciesPatterns types.List   `tfsdk:"external_dependencies_patterns"`
	EnableTokenAuthentication    types.Bool   `tfsdk:"enable_token_authentication"`
	ProjectID                    types.String `tfsdk:"project_id"`
}

func (r *remoteHelmOCIResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteHelmOCIResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHelmOCIResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHelmOCIResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHelmOCIResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteHelmOCIResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHelmOCIResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteHelmOCIResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
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

	return RemoteHelmOCIAPIModel{
		RemoteAPIModel:               remoteAPIModel,
		ExternalDependenciesEnabled:  r.ExternalDependenciesEnabled.ValueBool(),
		ExternalDependenciesPatterns: externalDependenciesPatterns,
		EnableTokenAuthentication:    r.EnableTokenAuthentication.ValueBool(),
		ProjectID:                    r.ProjectID.ValueString(),
	}, diags
}

func (r *remoteHelmOCIResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteHelmOCIAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.ExternalDependenciesEnabled = types.BoolValue(model.ExternalDependenciesEnabled)
	r.EnableTokenAuthentication = types.BoolValue(model.EnableTokenAuthentication)

	externalDependenciesPatterns, d := types.ListValueFrom(ctx, types.StringType, model.ExternalDependenciesPatterns)
	if d != nil {
		diags.Append(d...)
	}
	r.ExternalDependenciesPatterns = externalDependenciesPatterns
	r.ProjectID = types.StringValue(model.ProjectID)

	return diags
}

type RemoteHelmOCIAPIModel struct {
	RemoteAPIModel
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	ProjectID                    string   `json:"dockerProjectId"`
}

func (r *remoteHelmOCIResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteHelmAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"external_dependencies_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, external dependencies are rewritten. External Dependency Rewrite in the UI.",
			},
			"enable_token_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable token (Bearer) based authentication.",
			},
			"external_dependencies_patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.AlsoRequires(path.MatchRoot("external_dependencies_enabled")),
				},
				MarkdownDescription: "Optional include patterns to match external URLs. Ant-style path expressions are supported (*, **, ?). " +
					"For example, specifying `**/github.com/**` will only allow downloading foreign layers from github.com host." +
					"By default, this is set to '**' in the UI, which means that foreign layers may be downloaded from any external host." +
					"This attribute must be set together with `external_dependencies_enabled = true`",
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				MarkdownDescription: "Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific " +
					"project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id " +
					"will default to their default project as you have set up in your account.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteHelmAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
