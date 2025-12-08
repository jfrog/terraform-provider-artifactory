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

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewNugetLocalRepositoryResource() resource.Resource {
	return &localNugetResource{
		localResource: NewLocalRepositoryResource(
			repository.NugetPackageType,
			"NuGet",
			reflect.TypeFor[LocalNugetResourceModel](),
			reflect.TypeFor[LocalNugetAPIModel](),
		),
	}
}

type localNugetResource struct {
	localResource
}

type LocalNugetResourceModel struct {
	LocalResourceModel
	MaxUniqueSnapshots       types.Int64 `tfsdk:"max_unique_snapshots"`
	ForceNugetAuthentication types.Bool  `tfsdk:"force_nuget_authentication"`
}

func (r *LocalNugetResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalNugetResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalNugetResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalNugetResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalNugetResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalNugetResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalNugetResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalNugetResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalNugetAPIModel{
		LocalAPIModel:            localAPIModel,
		MaxUniqueSnapshots:       r.MaxUniqueSnapshots.ValueInt64(),
		ForceNugetAuthentication: r.ForceNugetAuthentication.ValueBool(),
	}, diags
}

func (r *LocalNugetResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalNugetAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.MaxUniqueSnapshots = types.Int64Value(model.MaxUniqueSnapshots)
	r.ForceNugetAuthentication = types.BoolValue(model.ForceNugetAuthentication)

	return diags
}

type LocalNugetAPIModel struct {
	LocalAPIModel
	MaxUniqueSnapshots       int64 `json:"maxUniqueSnapshots"`
	ForceNugetAuthentication bool  `json:"forceNugetAuthentication"`
}

func (r *localNugetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"max_unique_snapshots": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The maximum number of unique snapshots of a single artifact to store. Once the number of snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.",
			},
			"force_nuget_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force basic authentication credentials in order to use this repository.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var nugetSchema = lo.Assign(
	map[string]*sdkv2_schema.Schema{
		"max_unique_snapshots": {
			Type:     sdkv2_schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique snapshots of a single artifact to store.\nOnce the number of " +
				"snapshots exceeds this setting, older versions are removed.\nA value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.",
		},
		"force_nuget_authentication": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Force basic authentication credentials in order to use this repository.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.NugetPackageType),
)

var NugetSchemas = GetSchemas(nugetSchema)

type NugetLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

//
//
// func ResourceArtifactoryLocalNugetRepository() *schema.Resource {
//
// 	var unPackLocalNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
// 		repo := UnpackLocalNugetRepository(data, Rclass)
// 		return repo, repo.Id(), nil
// 	}
//
// 	constructor := func() (interface{}, error) {
// 		return &NugetLocalRepositoryParams{
// 			RepositoryBaseParams: RepositoryBaseParams{
// 				PackageType: repository.NugetPackageType,
// 				Rclass:      Rclass,
// 			},
// 			MaxUniqueSnapshots:       0,
// 			ForceNugetAuthentication: false,
// 		}, nil
// 	}
//
// 	return repository.MkResourceSchema(
// 		NugetSchemas,
// 		packer.Default(NugetSchemas[CurrentSchemaVersion]),
// 		unPackLocalNugetRepository,
// 		constructor,
// 	)
// }
