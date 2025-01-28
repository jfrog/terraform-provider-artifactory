package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

func NewPyPIRemoteRepositoryResource() resource.Resource {
	return &remotePyPIResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.PyPiPackageType,
			repository.PackageNameLookup[repository.PyPiPackageType],
			reflect.TypeFor[remotePyPIResourceModel](),
			reflect.TypeFor[RemotePyPIAPIModel](),
		),
	}
}

type remotePyPIResource struct {
	remoteResource
}

type remotePyPIResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	PyPIRegistryURL      types.String `tfsdk:"pypi_registry_url"`
	PyPIRepositorySuffix types.String `tfsdk:"pypi_repository_suffix"`
}

func (r *remotePyPIResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remotePyPIResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remotePyPIResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remotePyPIResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remotePyPIResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remotePyPIResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remotePyPIResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remotePyPIResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemotePyPIAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
		PyPIRegistryURL:      r.PyPIRegistryURL.ValueString(),
		PyPIRepositorySuffix: r.PyPIRepositorySuffix.ValueString(),
	}, diags
}

func (r *remotePyPIResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemotePyPIAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.PyPIRegistryURL = types.StringValue(model.PyPIRegistryURL)
	r.PyPIRepositorySuffix = types.StringValue(model.PyPIRepositorySuffix)
	return diags
}

type RemotePyPIAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	PyPIRegistryURL      string `json:"pyPIRegistryUrl"`
	PyPIRepositorySuffix string `json:"pyPIRepositorySuffix"`
}

func (r *remotePyPIResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remotePyPIAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"pypi_registry_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://pypi.org"),
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "To configure the remote repo to proxy public external PyPI repository, or a PyPI repository hosted on another Artifactory server. See JFrog Pypi documentation for the usage details. Default value is 'https://pypi.org'.",
			},
			"pypi_repository_suffix": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("simple"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "Usually should be left as a default for 'simple', unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI. Default value is 'simple'.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remotePyPIAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
