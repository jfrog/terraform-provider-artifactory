package local

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewAlpineLocalRepositoryResource() resource.Resource {
	return &localAlpineResource{
		localResource: NewLocalRepositoryResource(
			repository.AlpinePackageType,
			"Alpine",
			reflect.TypeFor[LocalAlpineResourceModel](),
			reflect.TypeFor[LocalAlpineAPIModel](),
		),
	}
}

type localAlpineResource struct {
	localResource
}

type LocalAlpineResourceModel struct {
	LocalResourceModel
	PrimaryKeyPairRef types.String `tfsdk:"primary_keypair_ref"`
}

func (r *LocalAlpineResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalAlpineResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalAlpineResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalAlpineResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalAlpineResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalAlpineResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalAlpineResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalAlpineResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalAlpineAPIModel{
		LocalAPIModel:     localAPIModel,
		PrimaryKeyPairRef: r.PrimaryKeyPairRef.ValueStringPointer(),
	}, diags
}

func (r *LocalAlpineResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalAlpineAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.PrimaryKeyPairRef = types.StringPointerValue(model.PrimaryKeyPairRef)

	return diags
}

type LocalAlpineAPIModel struct {
	LocalAPIModel
	PrimaryKeyPairRef *string `json:"primaryKeyPairRef"`
}

var AlpinePrimaryKeyPairRefAttribute = map[string]schema.Attribute{
	"primary_keypair_ref": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString(""),
		MarkdownDescription: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
	},
}

func (r *localAlpineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		AlpinePrimaryKeyPairRefAttribute,
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var alpineSchema = lo.Assign(
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AlpinePackageType),
	repository.AlpinePrimaryKeyPairRefSDKv2,
	repository.CompressionFormatsSDKv2,
)

var AlpineLocalSchemas = GetSchemas(alpineSchema)

type AlpineLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}
