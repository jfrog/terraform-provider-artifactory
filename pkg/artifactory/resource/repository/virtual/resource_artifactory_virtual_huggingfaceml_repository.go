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

package virtual

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewHuggingFaceMLVirtualRepositoryResource() resource.Resource {
	return &virtualHuggingFaceMLResource{
		virtualResource: NewVirtualRepositoryResource(
			repository.HuggingFacePackageType,
			repository.PackageNameLookup[repository.HuggingFacePackageType],
			reflect.TypeFor[VirtualHuggingFaceMLResourceModel](),
			reflect.TypeFor[VirtualHuggingFaceMLAPIModel](),
		),
	}
}

type virtualHuggingFaceMLResource struct {
	virtualResource
}

type VirtualHuggingFaceMLResourceModel struct {
	VirtualResourceModel
}

func (r *VirtualHuggingFaceMLResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r VirtualHuggingFaceMLResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *VirtualHuggingFaceMLResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r VirtualHuggingFaceMLResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *VirtualHuggingFaceMLResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *VirtualHuggingFaceMLResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r VirtualHuggingFaceMLResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r VirtualHuggingFaceMLResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	virtualAPIModel, d := r.VirtualResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return VirtualHuggingFaceMLAPIModel{
		VirtualAPIModel: virtualAPIModel.(VirtualAPIModel),
	}, diags
}

func (r *VirtualHuggingFaceMLResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*VirtualHuggingFaceMLAPIModel)

	d := r.VirtualResourceModel.FromAPIModel(ctx, model.VirtualAPIModel)
	if d != nil {
		diags.Append(d...)
	}

	return diags
}

type VirtualHuggingFaceMLAPIModel struct {
	VirtualAPIModel
}

func (r *virtualHuggingFaceMLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	virtualHuggingFaceMLAttributes := lo.Assign(
		VirtualAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
	)

	resp.Schema = schema.Schema{
		Version:     1,
		Attributes:  virtualHuggingFaceMLAttributes,
		Description: r.Description,
	}
}
