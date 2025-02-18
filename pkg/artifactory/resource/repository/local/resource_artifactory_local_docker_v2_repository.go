package local

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

func NewDockerV2LocalRepositoryResource() resource.Resource {
	return &localDockerV2Resource{
		localResource: localResource{
			BaseResource: repository.BaseResource{
				JFrogResource: util.JFrogResource{
					TypeName:           fmt.Sprintf("artifactory_local_%s_v2_repository", repository.DockerPackageType),
					CollectionEndpoint: "artifactory/api/repositories",
					DocumentEndpoint:   "artifactory/api/repositories/{key}",
				},
				Description:       "Provides a resource to creates a Docker V2 repository.",
				PackageType:       repository.DockerPackageType,
				Rclass:            Rclass,
				ResourceModelType: reflect.TypeFor[LocalDockerV2ResourceModel](),
				APIModelType:      reflect.TypeFor[LocalDockerV2APIModel](),
			},
		},
	}
}

type localDockerV2Resource struct {
	localResource
}

type LocalDockerV2ResourceModel struct {
	LocalResourceModel
	MaxUniqueTags       types.Int64  `tfsdk:"max_unique_tags"`
	TagRetention        types.Int64  `tfsdk:"tag_retention"`
	BlockPushingSchema1 types.Bool   `tfsdk:"block_pushing_schema1"`
	APIVersion          types.String `tfsdk:"api_version"`
}

func (r *LocalDockerV2ResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalDockerV2ResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV2ResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDockerV2ResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV2ResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalDockerV2ResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalDockerV2ResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalDockerV2ResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	r.APIVersion = types.StringValue("V2")

	return LocalDockerV2APIModel{
		LocalAPIModel:       localAPIModel,
		MaxUniqueTags:       r.MaxUniqueTags.ValueInt64(),
		TagRetention:        r.TagRetention.ValueInt64(),
		DockerApiVersion:    "V2",
		BlockPushingSchema1: r.BlockPushingSchema1.ValueBool(),
	}, diags
}

func (r *LocalDockerV2ResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalDockerV2APIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.MaxUniqueTags = types.Int64Value(model.MaxUniqueTags)
	r.TagRetention = types.Int64Value(model.TagRetention)
	r.BlockPushingSchema1 = types.BoolValue(model.BlockPushingSchema1)
	r.APIVersion = types.StringValue(model.DockerApiVersion)

	return diags
}

type LocalDockerV2APIModel struct {
	LocalAPIModel
	MaxUniqueTags       int64  `json:"maxUniqueTags"`
	DockerApiVersion    string `json:"dockerApiVersion"`
	TagRetention        int64  `json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `json:"blockPushingSchema1"`
}

func (r *localDockerV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"max_unique_tags": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The maximum number of unique tags of a single Docker image to store in this repository.\n" +
					"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.\n" +
					"This only applies to manifest v2",
			},
			"tag_retention": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "If greater than 1, overwritten tags will be saved by their digest, up to the set up number.",
			},
			"block_pushing_schema1": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, Artifactory will block the pushing of Docker images with manifest v2 schema 1 to this repository.",
			},
			"api_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Docker API version to use.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

type DockerLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags"`
	DockerApiVersion    string `hcl:"api_version" json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

var dockerV2Schema = lo.Assign(
	map[string]*sdkv2_schema.Schema{
		"max_unique_tags": {
			Type:     sdkv2_schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique tags of a single Docker image to store in this repository.\n" +
				"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.\n" +
				"This only applies to manifest v2",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"tag_retention": {
			Type:             sdkv2_schema.TypeInt,
			Optional:         true,
			Computed:         false,
			Description:      "If greater than 1, overwritten tags will be saved by their digest, up to the set up number. This only applies to manifest V2",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
		},
		"block_pushing_schema1": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "When set, Artifactory will block the pushing of Docker images with manifest v2 schema 1 to this repository.",
		},
		"api_version": {
			Type:        sdkv2_schema.TypeString,
			Computed:    true,
			Description: "The Docker API version to use. This cannot be set",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
)

var DockerV2Schemas = GetSchemas(dockerV2Schema)
