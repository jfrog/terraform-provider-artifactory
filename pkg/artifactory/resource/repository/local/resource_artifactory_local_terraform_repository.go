package local

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewTerraformModuleLocalRepositoryResource() resource.Resource {
	return &localTerraformResource{
		localResource: NewLocalRepositoryResource(
			repository.TerraformModulePackageType,
			"Terraform Module",
			reflect.TypeFor[LocalTerraformResourceModel](),
			reflect.TypeFor[LocalTerraformAPIModel](),
		),
	}
}

func NewTerraformProviderLocalRepositoryResource() resource.Resource {
	return &localTerraformResource{
		localResource: NewLocalRepositoryResource(
			repository.TerraformProviderPackageType,
			"Terraform Provider",
			reflect.TypeFor[LocalTerraformResourceModel](),
			reflect.TypeFor[LocalTerraformAPIModel](),
		),
	}
}

type localTerraformResource struct {
	localResource
}

type LocalTerraformResourceModel struct {
	LocalResourceModel
}

func (r *LocalTerraformResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalTerraformResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalTerraformResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalTerraformResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalTerraformResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalTerraformResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalTerraformResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalTerraformResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()
	localAPIModel.PackageType = repository.TerraformPackageType

	var terraformType string
	switch packageType {
	case repository.TerraformModulePackageType:
		terraformType = "module"
	case repository.TerraformProviderPackageType:
		terraformType = "provider"
	}

	return LocalTerraformAPIModel{
		LocalAPIModel: localAPIModel,
		TerraformType: terraformType,
	}, diags
}

func (r *LocalTerraformResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalTerraformAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)

	return diags
}

type LocalTerraformAPIModel struct {
	LocalAPIModel
	TerraformType string `json:"terraformType"`
}

func (r *localTerraformResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

func GetTerraformSchemas(registryType string) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, "terraform_"+registryType),
		),
		1: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, "terraform_"+registryType),
		),
	}
}
