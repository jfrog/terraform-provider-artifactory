// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func NewCargoLocalRepositoryResource() resource.Resource {
	return &localCargoResource{
		localResource: NewLocalRepositoryResource(
			repository.CargoPackageType,
			"Cargo",
			reflect.TypeFor[LocalCargoResourceModel](),
			reflect.TypeFor[LocalCargoAPIModel](),
		),
	}
}

type localCargoResource struct {
	localResource
}

type LocalCargoResourceModel struct {
	LocalResourceModel
	AnonymousAccess   types.Bool `tfsdk:"anonymous_access"`
	EnableSparseIndex types.Bool `tfsdk:"enable_sparse_index"`
}

func (r *LocalCargoResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalCargoResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalCargoResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalCargoResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalCargoResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalCargoResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalCargoResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalCargoResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalCargoAPIModel{
		LocalAPIModel:     localAPIModel,
		AnonymousAccess:   r.AnonymousAccess.ValueBool(),
		EnableSparseIndex: r.EnableSparseIndex.ValueBool(),
	}, diags
}

func (r *LocalCargoResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalCargoAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.AnonymousAccess = types.BoolValue(model.AnonymousAccess)
	r.EnableSparseIndex = types.BoolValue(model.EnableSparseIndex)

	return diags
}

type LocalCargoAPIModel struct {
	LocalAPIModel
	AnonymousAccess   bool `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool `json:"cargoInternalIndex"`
}

func (r *localCargoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"anonymous_access": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is `false`.",
			},
			"enable_sparse_index": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is `false`.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var cargoSchema = lo.Assign(
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.CargoPackageType),
	map[string]*sdkv2_schema.Schema{
		"anonymous_access": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is 'false'.",
		},
		"enable_sparse_index": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is 'false'.",
		},
	},
	repository.CompressionFormatsSDKv2,
)

var CargoSchemas = GetSchemas(cargoSchema)

type CargoLocalRepoParams struct {
	RepositoryBaseParams
	AnonymousAccess   bool `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool `json:"cargoInternalIndex"`
}
