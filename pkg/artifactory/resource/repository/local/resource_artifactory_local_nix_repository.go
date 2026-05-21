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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewNixLocalRepositoryResource() resource.Resource {
	return &localNixResource{
		localResource: NewLocalRepositoryResource(
			repository.NixPackageType,
			repository.PackageNameLookup[repository.NixPackageType],
			reflect.TypeFor[LocalNixResourceModel](),
			reflect.TypeFor[LocalNixAPIModel](),
		),
	}
}

type localNixResource struct {
	localResource
}

type LocalNixResourceModel struct {
	LocalResourceModel
}

func (r *LocalNixResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalNixResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalNixResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalNixResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalNixResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalNixResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalNixResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalNixResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalNixAPIModel{
		LocalAPIModel: localAPIModel,
	}, diags
}

func (r *LocalNixResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalNixAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)

	return diags
}

type LocalNixAPIModel struct {
	LocalAPIModel
}

func (r *localNixResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}
