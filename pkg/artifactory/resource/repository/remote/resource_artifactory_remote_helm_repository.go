package remote

import (
	"context"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewHelmRemoteRepositoryResource() resource.Resource {
	return &remoteHelmResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.HelmPackageType,
			repository.PackageNameLookup[repository.HelmPackageType],
			reflect.TypeFor[remoteHelmResourceModel](),
			reflect.TypeFor[RemoteHelmAPIModel](),
		),
	}
}

type remoteHelmResource struct {
	remoteResource
}

type remoteHelmResourceModel struct {
	RemoteResourceModel
	HelmChartsBaseURL            types.String `tfsdk:"helm_charts_base_url"`
	ExternalDependenciesEnabled  types.Bool   `tfsdk:"external_dependencies_enabled"`
	ExternalDependenciesPatterns types.Set    `tfsdk:"external_dependencies_patterns"`
}

func (r *remoteHelmResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteHelmResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHelmResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHelmResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHelmResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteHelmResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHelmResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteHelmResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
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

	return RemoteHelmAPIModel{
		RemoteAPIModel:               remoteAPIModel,
		HelmChartsBaseURL:            r.HelmChartsBaseURL.ValueString(),
		ExternalDependenciesEnabled:  r.ExternalDependenciesEnabled.ValueBool(),
		ExternalDependenciesPatterns: externalDependenciesPatterns,
	}, diags
}

func (r *remoteHelmResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteHelmAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.HelmChartsBaseURL = types.StringValue(model.HelmChartsBaseURL)
	r.ExternalDependenciesEnabled = types.BoolValue(model.ExternalDependenciesEnabled)

	externalDependenciesPatterns, d := types.SetValueFrom(ctx, types.StringType, model.ExternalDependenciesPatterns)
	if d != nil {
		diags.Append(d...)
	}
	r.ExternalDependenciesPatterns = externalDependenciesPatterns

	return diags
}

type RemoteHelmAPIModel struct {
	RemoteAPIModel
	HelmChartsBaseURL            string   `json:"chartsBaseUrl"`
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns"`
}

func (r *remoteHelmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteHelmAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"helm_charts_base_url": schema.StringAttribute{
				Optional: true,
				// Computed: true,
				// Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^(?:http|https|oci):\/\/.+$`), "must start with http, https, or oci"),
				},
				MarkdownDescription: "Base URL for the translation of chart source URLs in the index.yaml of virtual repos. " +
					"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url. " +
					"Support http/https/oci protocol scheme.",
			},
			"external_dependencies_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, external dependencies are rewritten. External Dependency Rewrite in the UI.",
			},
			"external_dependencies_patterns": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.AlsoRequires(path.MatchRoot("external_dependencies_enabled")),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
					"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response." +
					"Default value in UI is empty. This attribute must be set together with `external_dependencies_enabled = true`",
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
