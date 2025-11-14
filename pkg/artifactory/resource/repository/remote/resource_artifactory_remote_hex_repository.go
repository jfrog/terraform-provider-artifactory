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

func NewHexRemoteRepositoryResource() resource.Resource {
	return &remoteHexResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.HexPackageType,
			repository.PackageNameLookup[repository.HexPackageType],
			reflect.TypeFor[remoteHexResourceModel](),
			reflect.TypeFor[RemoteHexAPIModel](),
		),
	}
}

type remoteHexResource struct {
	remoteResource
}

type remoteHexResourceModel struct {
	RemoteResourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
	PublicKey            types.String `tfsdk:"public_key"`
}

func (r *remoteHexResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteHexResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHexResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHexResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteHexResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteHexResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteHexResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteHexResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteHexAPIModel{
		RemoteAPIModel:       remoteAPIModel,
		HexPrimaryKeyPairRef: r.HexPrimaryKeyPairRef.ValueString(),
		PublicKey:            r.PublicKey.ValueString(),
	}, diags
}

func (r *remoteHexResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteHexAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.HexPrimaryKeyPairRef = types.StringValue(model.HexPrimaryKeyPairRef)
	r.PublicKey = types.StringValue(model.PublicKey)

	return diags
}

type RemoteHexAPIModel struct {
	RemoteAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
	PublicKey            string `json:"hexPublicKey"`
}

func (r *remoteHexResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteHexAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"hex_primary_keypair_ref": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
			},
			"public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteHexAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
