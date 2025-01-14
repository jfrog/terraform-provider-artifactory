package local

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

func NewDockerV1LocalRepositoryResource() resource.Resource {
	return &localDockerV1Resource{
		localResource: localResource{
			BaseResource: repository.BaseResource{
				JFrogResource: util.JFrogResource{
					TypeName:           fmt.Sprintf("artifactory_local_%s_v1_repository", repository.DockerPackageType),
					CollectionEndpoint: "artifactory/api/repositories",
					DocumentEndpoint:   "artifactory/api/repositories/{key}",
				},
				Description:       "Provides a resource to creates a Docker V1 repository.",
				PackageType:       repository.DockerPackageType,
				Rclass:            Rclass,
				ResourceModelType: reflect.TypeFor[LocalDockerV1ResourceModel](),
				APIModelType:      reflect.TypeFor[LocalDockerV1APIModel](),
			},
		},
	}
}

type localDockerV1Resource struct {
	localResource
}

type LocalDockerV1ResourceModel struct {
	LocalResourceModel
	MaxUniqueTags       types.Int64  `tfsdk:"max_unique_tags"`
	TagRetention        types.Int64  `tfsdk:"tag_retention"`
	BlockPushingSchema1 types.Bool   `tfsdk:"block_pushing_schema1"`
	APIVersion          types.String `tfsdk:"api_version"`
}

func (r *LocalDockerV1ResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalDockerV1ResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV1ResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDockerV1ResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV1ResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalDockerV1ResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDockerV1ResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV1ResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	r.MaxUniqueTags = types.Int64Value(0)
	r.TagRetention = types.Int64Value(1)
	r.APIVersion = types.StringValue("V1")
	r.BlockPushingSchema1 = types.BoolValue(false)

	return LocalDockerV1APIModel{
		LocalAPIModel:       localAPIModel,
		MaxUniqueTags:       0,
		TagRetention:        1,
		DockerApiVersion:    "V1",
		BlockPushingSchema1: false,
	}, diags
}

func (r *LocalDockerV1ResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalDockerV1APIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.MaxUniqueTags = types.Int64Value(model.MaxUniqueTags)
	r.TagRetention = types.Int64Value(model.TagRetention)
	r.BlockPushingSchema1 = types.BoolValue(model.BlockPushingSchema1)
	r.APIVersion = types.StringValue(model.DockerApiVersion)

	return diags
}

type LocalDockerV1APIModel struct {
	LocalAPIModel
	MaxUniqueTags       int64  `json:"maxUniqueTags"`
	DockerApiVersion    string `json:"dockerApiVersion"`
	TagRetention        int64  `json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `json:"blockPushingSchema1"`
}

func (r *localDockerV1Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"max_unique_tags": schema.Int64Attribute{
				Computed: true,
			},
			"tag_retention": schema.Int64Attribute{
				Computed: true,
			},
			"block_pushing_schema1": schema.BoolAttribute{
				Computed: true,
			},
			"api_version": schema.StringAttribute{
				Computed: true,
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var dockerV1Schema = utilsdk.MergeMaps(
	map[string]*sdkv2_schema.Schema{
		"max_unique_tags": {
			Type:     sdkv2_schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"tag_retention": {
			Type:     sdkv2_schema.TypeInt,
			Computed: true,
		},
		"block_pushing_schema1": {
			Type:     sdkv2_schema.TypeBool,
			Computed: true,
		},
		"api_version": {
			Type:     sdkv2_schema.TypeString,
			Computed: true,
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
)

var DockerV1Schemas = GetSchemas(dockerV1Schema)
