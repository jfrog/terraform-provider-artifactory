package virtual

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewHexVirtualRepositoryResource() resource.Resource {
	return &virtualHexResource{
		BaseResource: repository.NewRepositoryResource(
			repository.HexPackageType,
			"hex",
			"virtual",
			reflect.TypeOf(virtualHexResourceModel{}),
			reflect.TypeOf(VirtualHexAPIModel{}),
		),
	}
}

type virtualHexResource struct {
	repository.BaseResource
}

type virtualHexResourceModel struct {
	VirtualResourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
}

// Resource lifecycle methods for virtualHexResourceModel
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
	// Read Terraform state data into the model
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

func (r virtualHexResourceModel) KeyString() string {
	return r.VirtualResourceModel.BaseResourceModel.KeyString()
}

func (r virtualHexResourceModel) ProjectKeyValue() basetypes.StringValue {
	return r.VirtualResourceModel.BaseResourceModel.ProjectKeyValue()
}

func (r virtualHexResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Get virtual API model
	virtualModel, d := r.VirtualResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return VirtualHexAPIModel{
		VirtualAPIModel:      virtualModel.(VirtualAPIModel),
		HexPrimaryKeyPairRef: r.HexPrimaryKeyPairRef.ValueString(),
	}, diags
}

func (r *virtualHexResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*VirtualHexAPIModel)

	// Set virtual model fields
	r.VirtualResourceModel.FromAPIModel(ctx, model.VirtualAPIModel)
	r.HexPrimaryKeyPairRef = types.StringValue(model.HexPrimaryKeyPairRef)

	return diags
}

type VirtualHexAPIModel struct {
	VirtualAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (r *virtualHexResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Combine virtual attributes with Hex-specific attributes
	hexAttributes := lo.Assign(
		VirtualAttributes,
		repository.RepoLayoutRefAttribute("virtual", repository.HexPackageType),
		map[string]schema.Attribute{
			"hex_primary_keypair_ref": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Provides a resource to create a virtual Hex repository.",
		Attributes:  hexAttributes,
	}
}
