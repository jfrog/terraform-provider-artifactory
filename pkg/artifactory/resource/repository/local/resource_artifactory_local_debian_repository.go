package local

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewDebianLocalRepositoryResource() resource.Resource {
	return &localDebianResource{
		localResource: NewLocalRepositoryResource(
			repository.DebianPackageType,
			"Debian",
			reflect.TypeFor[LocalDebianResourceModel](),
			reflect.TypeFor[LocalDebianAPIModel](),
		),
	}
}

type localDebianResource struct {
	localResource
}

type LocalDebianResourceModel struct {
	LocalResourceModel
	PrimaryKeyPairRef   types.String `tfsdk:"primary_keypair_ref"`
	SecondaryKeyPairRef types.String `tfsdk:"secondary_keypair_ref"`
	CompressionFormats  types.Set    `tfsdk:"index_compression_formats"`
	TrivialLayout       types.Bool   `tfsdk:"trivial_layout"`
	DdebSupported       types.Bool   `tfsdk:"ddeb_supported"`
}

func (r *LocalDebianResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalDebianResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDebianResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDebianResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDebianResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalDebianResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDebianResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalDebianResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	var compressionFormats []string
	d = r.CompressionFormats.ElementsAs(ctx, &compressionFormats, false)
	if d != nil {
		diags.Append(d...)
	}

	return LocalDebianAPIModel{
		LocalAPIModel:       localAPIModel,
		PrimaryKeyPairRef:   r.PrimaryKeyPairRef.ValueString(),
		SecondaryKeyPairRef: r.SecondaryKeyPairRef.ValueString(),
		CompressionFormats:  compressionFormats,
		TrivialLayout:       r.TrivialLayout.ValueBool(),
		DdebSupported:       r.DdebSupported.ValueBool(),
	}, diags
}

func (r *LocalDebianResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalDebianAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.PrimaryKeyPairRef = types.StringValue(model.PrimaryKeyPairRef)
	r.SecondaryKeyPairRef = types.StringValue(model.SecondaryKeyPairRef)
	compressionFormats, d := types.SetValueFrom(ctx, types.StringType, model.CompressionFormats)
	if d != nil {
		diags.Append(d...)
	}

	r.CompressionFormats = compressionFormats
	r.TrivialLayout = types.BoolValue(model.TrivialLayout)
	r.DdebSupported = types.BoolValue(model.DdebSupported)

	return diags
}

type LocalDebianAPIModel struct {
	LocalAPIModel
	PrimaryKeyPairRef   string   `json:"primaryKeyPairRef"`
	SecondaryKeyPairRef string   `json:"secondaryKeyPairRef"`
	CompressionFormats  []string `json:"optionalIndexCompressionFormats,omitempty"`
	TrivialLayout       bool     `json:"debianTrivialLayout"`
	DdebSupported       bool     `json:"ddebSupported"`
}

func (r *localDebianResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		repository.CompressionFormatsAttribute,
		repository.PrimaryKeyPairRefAttribute,
		repository.SecondaryKeyPairRefAttribute,
		map[string]schema.Attribute{
			"trivial_layout": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, the repository will use the deprecated trivial layout.",
			},
			"ddeb_supported": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, enable indexing with debug symbols (.ddeb).",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var debianSchema = lo.Assign(
	repository.PrimaryKeyPairRefSDKv2,
	repository.SecondaryKeyPairRefSDKv2,
	map[string]*sdkv2_schema.Schema{
		"trivial_layout": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, the repository will use the deprecated trivial layout.",
			Deprecated:  "You shouldn't be using this",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DebianPackageType),
	repository.CompressionFormatsSDKv2,
)

var DebianSchemas = GetSchemas(debianSchema)

type DebianLocalRepositoryParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
	repository.SecondaryKeyPairRefParam
	TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout"`
	IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
}
