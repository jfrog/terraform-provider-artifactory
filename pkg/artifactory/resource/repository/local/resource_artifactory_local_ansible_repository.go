package local

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

func NewAnsibleLocalRepositoryResource() resource.Resource {
	return &localAnsibleResource{
		localResource: NewLocalRepositoryResource(repository.AnsiblePackageType, reflect.TypeFor[LocalAnsibleResourceModel](), reflect.TypeFor[LocalAnsibleAPIModel]()),
	}
}

type localAnsibleResource struct {
	localResource
}

type LocalAnsibleResourceModel struct {
	LocalResourceModel
	PrimaryKeyPairRef types.String `tfsdk:"primary_keypair_ref"`
}

func (r *LocalAnsibleResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalAnsibleResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalAnsibleResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalAnsibleResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalAnsibleResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalAnsibleResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalAnsibleResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalAnsibleResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalAnsibleAPIModel{
		LocalAPIModel:     localAPIModel,
		PrimaryKeyPairRef: r.PrimaryKeyPairRef.ValueStringPointer(),
	}, diags
}

func (r *LocalAnsibleResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalAnsibleAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.PrimaryKeyPairRef = types.StringPointerValue(model.PrimaryKeyPairRef)

	return diags
}

type LocalAnsibleAPIModel struct {
	LocalAPIModel
	PrimaryKeyPairRef *string `json:"primaryKeyPairRef"`
}

func (r *localAnsibleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		repository.PrimaryKeyPairRefAttribute,
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var ansibleSchema = lo.Assign(
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AnsiblePackageType),
	repository.PrimaryKeyPairRefSDKv2,
)

var AnsibleSchemas = GetSchemas(ansibleSchema)

type AnsibleLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}
