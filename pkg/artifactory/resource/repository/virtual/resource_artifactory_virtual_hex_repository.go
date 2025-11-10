package virtual

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

func NewHexVirtualRepositoryResource() resource.Resource {
	return &virtualHexResource{
		virtualResource: NewVirtualRepositoryResource(
			repository.HexPackageType,
			repository.PackageNameLookup[repository.HexPackageType],
			reflect.TypeFor[virtualHexResourceModel](),
			reflect.TypeFor[VirtualHexAPIModel](),
		),
	}
}

type virtualHexResource struct {
	virtualResource
}

type virtualHexResourceModel struct {
	VirtualResourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
}

func (r *virtualHexResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r virtualHexResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *virtualHexResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r virtualHexResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *virtualHexResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *virtualHexResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r virtualHexResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r virtualHexResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	virtualAPIModel, d := r.VirtualResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return VirtualHexAPIModel{
		VirtualAPIModel:      virtualAPIModel.(VirtualAPIModel),
		HexPrimaryKeyPairRef: r.HexPrimaryKeyPairRef.ValueString(),
	}, diags
}

func (r *virtualHexResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*VirtualHexAPIModel)

	r.VirtualResourceModel.FromAPIModel(ctx, model.VirtualAPIModel)
	r.HexPrimaryKeyPairRef = types.StringValue(model.HexPrimaryKeyPairRef)

	return diags
}

type VirtualHexAPIModel struct {
	VirtualAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (r *virtualHexResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	virtualHexAttributes := lo.Assign(
		VirtualAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"hex_primary_keypair_ref": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     1,
		Attributes:  virtualHexAttributes,
		Description: r.Description,
	}
}
