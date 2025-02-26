package local

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewHelmLocalRepositoryResource() resource.Resource {
	return &localHelmResource{
		localResource: NewLocalRepositoryResource(
			repository.HelmPackageType,
			"Helm",
			reflect.TypeFor[LocalHelmResourceModel](),
			reflect.TypeFor[LocalHelmAPIModel](),
		),
	}
}

type localHelmResource struct {
	localResource
}

type LocalHelmResourceModel struct {
	LocalResourceModel
	ForceNonDuplicateChart   types.Bool `tfsdk:"force_non_duplicate_chart"`
	ForceMetadataNameVersion types.Bool `tfsdk:"force_metadata_name_version"`
}

func (r *LocalHelmResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalHelmResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalHelmResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalHelmResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalHelmResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalHelmResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalHelmResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalHelmResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalHelmAPIModel{
		LocalAPIModel:            localAPIModel,
		ForceNonDuplicateChart:   r.ForceNonDuplicateChart.ValueBoolPointer(),
		ForceMetadataNameVersion: r.ForceMetadataNameVersion.ValueBoolPointer(),
	}, diags
}

func (r *LocalHelmResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalHelmAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.ForceNonDuplicateChart = types.BoolPointerValue(model.ForceNonDuplicateChart)
	r.ForceMetadataNameVersion = types.BoolPointerValue(model.ForceMetadataNameVersion)

	return diags
}

type LocalHelmAPIModel struct {
	LocalAPIModel
	ForceNonDuplicateChart   *bool `json:"forceNonDuplicateChart,omitempty"`
	ForceMetadataNameVersion *bool `json:"forceMetadataNameVersion,omitempty"`
}

func (r *localHelmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"force_non_duplicate_chart": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					PreventUpdateModifier(),
				},
				MarkdownDescription: "Prevents the deployment of charts with the same name and version in different repository paths. Only available for 7.104.5 onward. Cannot be updated after it is set.",
			},
			"force_metadata_name_version": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					PreventUpdateModifier(),
				},
				MarkdownDescription: "Ensures that the chart name and version in the file name match the values in Chart.yaml and adhere to SemVer standards. Only available for 7.104.5 onward. Cannot be updated after it is set.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

// Plan Modifier

func PreventUpdateModifier() planmodifier.Bool {
	return preventUpdateModifier{}
}

// preventUpdateModifier implements the plan modifier.
type preventUpdateModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m preventUpdateModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute cannot change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m preventUpdateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyBool implements the plan modification logic.
func (m preventUpdateModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is no state value, e.g. creating first time
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is planned and state values are identical.
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Attribute cannot be updated",
		fmt.Sprintf("%s cannot be updated after it is set.", req.Path.String()),
	)
}
