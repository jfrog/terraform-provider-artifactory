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

const currentGemsSchemaVersion = 4

func NewGemsRemoteRepositoryResource() resource.Resource {
	return &remoteGemsResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GemsPackageType,
			repository.PackageNameLookup[repository.GemsPackageType],
			reflect.TypeFor[remoteGemsResourceModel](),
			reflect.TypeFor[RemoteGemsAPIModel](),
		),
	}
}

type remoteGemsResource struct {
	remoteResource
}

type remoteGemsResourceModel struct {
	RemoteGenericResourceModelV4
	CurationResourceModel
}

func (r *remoteGemsResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGemsResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGemsResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteGemsResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGemsResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteGemsResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteGenericResourceModelV4.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteGemsAPIModel{
		RemoteGenericAPIModel: remoteAPIModel.(RemoteGenericAPIModel),
		CurationAPIModel: CurationAPIModel{
			Curated:     r.Curated.ValueBool(),
			PassThrough: r.PassThrough.ValueBool(),
		},
	}, diags
}

func (r *remoteGemsResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteGemsAPIModel)

	r.RemoteGenericResourceModelV4.FromAPIModel(ctx, &model.RemoteGenericAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.PassThrough = types.BoolValue(model.CurationAPIModel.PassThrough)

	return diags
}

type RemoteGemsAPIModel struct {
	RemoteGenericAPIModel
	CurationAPIModel
}

func (r *remoteGemsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteGemsAttributes := lo.Assign(
		remoteGenericAttributesV4,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
	)

	resp.Schema = schema.Schema{
		Version:     currentGemsSchemaVersion,
		Attributes:  remoteGemsAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
