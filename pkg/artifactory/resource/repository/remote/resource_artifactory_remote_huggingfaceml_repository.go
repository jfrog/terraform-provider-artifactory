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
	"github.com/samber/lo"
)

func NewHuggingFaceMLRemoteRepositoryResource() resource.Resource {
	return &remoteHuggingFaceMLResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.HuggingFacePackageType,
			repository.PackageNameLookup[repository.HuggingFacePackageType],
			reflect.TypeFor[remoteHuggingFaceMLResourceModel](),
			reflect.TypeFor[RemoteHuggingFaceMLAPIModel](),
		),
	}
}

type remoteHuggingFaceMLResource struct {
	remoteResource
}

type remoteHuggingFaceMLResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
}

func (r *remoteHuggingFaceMLResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteHuggingFaceMLResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHuggingFaceMLResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHuggingFaceMLResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHuggingFaceMLResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteHuggingFaceMLResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHuggingFaceMLResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteHuggingFaceMLResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteHuggingFaceMLAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
	}, diags
}

func (r *remoteHuggingFaceMLResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteHuggingFaceMLAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	return diags
}

type RemoteHuggingFaceMLAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
}

func (r *remoteHuggingFaceMLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteHuggingFaceMLAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://huggingface.co"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "The remote repo URL. Default to 'https://huggingface.co'",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteHuggingFaceMLAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
