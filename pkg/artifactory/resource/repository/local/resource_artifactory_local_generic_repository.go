package local

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var PackageTypesLikeGeneric = []string{
	repository.BowerPackageType,
	repository.ChefPackageType,
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GitLFSPackageType,
	repository.GoPackageType,
	repository.HelmPackageType,
	repository.HuggingFacePackageType,
	repository.NPMPackageType,
	repository.OpkgPackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.SwiftPackageType,
	repository.TerraformBackendPackageType,
	repository.VagrantPackageType,
}

func NewGenericLocalRepositoryResource(packageType string) func() resource.Resource {
	return func() resource.Resource {
		return &localGenericResource{
			localResource: localResource{
				BaseResource: repository.BaseResource{
					JFrogResource: util.JFrogResource{
						TypeName:           fmt.Sprintf("artifactory_local_%s_repository", packageType),
						CollectionEndpoint: "artifactory/api/repositories",
						DocumentEndpoint:   "artifactory/api/repositories/{key}",
					},
					Description:       "Provides a resource to creates a local Machine Learning repository.",
					PackageType:       packageType,
					Rclass:            Rclass,
					ResourceModelType: reflect.TypeFor[LocalGenericResourceModel](),
					APIModelType:      reflect.TypeFor[LocalGenericAPIModel](),
				},
			},
		}
	}
}

type localGenericResource struct {
	localResource
}

type LocalGenericResourceModel struct {
	LocalResourceModel
	RepoLayoutRef types.String `tfsdk:"repo_layout_ref"`
	CDNRedirect   types.Bool   `tfsdk:"cdn_redirect"`
}

func (r *LocalGenericResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalGenericResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalGenericResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalGenericResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalGenericResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalGenericResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalGenericResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalGenericResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalGenericAPIModel{
		LocalAPIModel: localAPIModel,
		CDNRedirect:   r.CDNRedirect.ValueBoolPointer(),
	}, diags
}

func (r *LocalGenericResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalGenericAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.CDNRedirect = types.BoolPointerValue(model.CDNRedirect)

	return diags
}

type LocalGenericAPIModel struct {
	LocalAPIModel
	CDNRedirect *bool `json:"cdnRedirect"`
}

func (r *localGenericResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	localGenericAttributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"cdn_redirect": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is `false`",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     1,
		Attributes:  localGenericAttributes,
		Description: r.Description,
	}
}

func GetGenericSchemas(packageType string) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
		),
		1: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
		),
	}
}
