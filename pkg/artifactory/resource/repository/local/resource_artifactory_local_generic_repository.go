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
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
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
			localResource: NewLocalRepositoryResource(
				packageType,
				repository.PackageNameLookup[packageType],
				reflect.TypeFor[LocalGenericResourceModel](),
				reflect.TypeFor[LocalGenericAPIModel](),
			),
		}
	}
}

type localGenericResource struct {
	localResource
}

type LocalGenericResourceModel struct {
	LocalResourceModel
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
	}, diags
}

func (r *LocalGenericResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalGenericAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)

	return diags
}

type LocalGenericAPIModel struct {
	LocalAPIModel
}

func (r *localGenericResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	localGenericAttributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
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
