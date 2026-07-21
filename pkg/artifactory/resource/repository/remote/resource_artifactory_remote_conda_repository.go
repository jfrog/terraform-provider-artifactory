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

package remote

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

func NewCondaRemoteRepositoryResource() resource.Resource {
	return &remoteCondaResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.CondaPackageType,
			repository.PackageNameLookup[repository.CondaPackageType],
			reflect.TypeFor[remoteCondaResourceModel](),
			reflect.TypeFor[RemoteCondaAPIModel](),
		),
	}
}

type remoteCondaResource struct {
	remoteResource
}

type remoteCondaResourceModel struct {
	RemoteGenericResourceModelV4
	CurationResourceModel
}

func (r *remoteCondaResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteCondaResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCondaResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCondaResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCondaResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteCondaResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCondaResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteCondaResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteGenericResourceModelV4.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteCondaAPIModel{
		RemoteGenericAPIModel: remoteAPIModel.(RemoteGenericAPIModel),
		CurationAPIModel: CurationAPIModel{
			Curated:     r.Curated.ValueBool(),
			PassThrough: r.PassThrough.ValueBool(),
		},
	}, diags
}

func (r *remoteCondaResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteCondaAPIModel)

	r.RemoteGenericResourceModelV4.FromAPIModel(ctx, &model.RemoteGenericAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.PassThrough = types.BoolValue(model.CurationAPIModel.PassThrough)

	return diags
}

type RemoteCondaAPIModel struct {
	RemoteGenericAPIModel
	CurationAPIModel
}

func (r *remoteCondaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteCondaAttributes := lo.Assign(
		remoteGenericAttributesV4,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteCondaAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
