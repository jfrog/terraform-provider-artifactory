package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

const currentGemsSchemaVersion = 4

func NewGemsRemoteRepositoryResource() resource.Resource {
	return &remoteGemsResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GemsPackageType,
			repository.PackageNameLookup[repository.GemsPackageType],
			reflect.TypeFor[remoteGemsResourceModel](),
			reflect.TypeFor[RemoteGemsAPIModel](),
		),
	}
}

type remoteGemsResource struct {
	remoteResource
}

type remoteGemsResourceModel struct {
	RemoteGenericResourceModelV4
	CurationResourceModel
}

func (r *remoteGemsResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGemsResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGemsResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteGemsResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteGemsResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteGenericResourceModelV4.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteGemsAPIModel{
		RemoteGenericAPIModel: remoteAPIModel.(RemoteGenericAPIModel),
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
	}, diags
}

func (r *remoteGemsResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteGemsAPIModel)

	r.RemoteGenericResourceModelV4.FromAPIModel(ctx, &model.RemoteGenericAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)

	return diags
}

type RemoteGemsAPIModel struct {
	RemoteGenericAPIModel
	CurationAPIModel
}

func (r *remoteGemsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteGemsAttributes := lo.Assign(
		remoteGenericAttributesV4,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
	)

	resp.Schema = schema.Schema{
		Version:     currentGemsSchemaVersion,
		Attributes:  remoteGemsAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
