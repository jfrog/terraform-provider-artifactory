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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewGenericRemoteRepositoryResource() resource.Resource {
	return &remoteGenericResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GenericPackageType,
			repository.PackageNameLookup[repository.GenericPackageType],
			reflect.TypeFor[RemoteGenericResourceModelV4](),
			reflect.TypeFor[RemoteGenericAPIModel](),
		),
	}
}

type remoteGenericResource struct {
	remoteResource
}

type RemoteGenericResourceModelV2 struct {
	RemoteResourceModel
}

type RemoteGenericResourceModelV3 struct {
	RemoteGenericResourceModelV2
	PropagateQueryParams types.Bool `tfsdk:"propagate_query_params"`
}

type RemoteGenericResourceModelV4 struct {
	RemoteGenericResourceModelV3
	RetrieveSha256FromServer types.Bool `tfsdk:"retrieve_sha256_from_server"`
}

func (r *RemoteGenericResourceModelV4) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV4) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteGenericResourceModelV4) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteGenericResourceModelV4) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteGenericResourceModelV4) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV4) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteGenericResourceModelV4) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r RemoteGenericResourceModelV4) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	if !r.RepoLayoutRef.IsNull() {
		remoteAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()
	}

	return RemoteGenericAPIModel{
		RemoteAPIModel:           remoteAPIModel,
		PropagateQueryParams:     r.PropagateQueryParams.ValueBool(),
		RetrieveSha256FromServer: r.RetrieveSha256FromServer.ValueBool(),
	}, diags
}

func (r *RemoteGenericResourceModelV4) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteGenericAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.PropagateQueryParams = types.BoolValue(model.PropagateQueryParams)
	r.RetrieveSha256FromServer = types.BoolValue(model.RetrieveSha256FromServer)

	return diags
}

type RemoteGenericAPIModel struct {
	RemoteAPIModel
	PropagateQueryParams     bool `json:"propagateQueryParams"`
	RetrieveSha256FromServer bool `json:"retrieveSha256FromServer"`
}

var remoteGenericAttributesV2 = lo.Assign(
	RemoteAttributes,
	repository.RepoLayoutRefAttribute(Rclass, repository.GenericPackageType),
)

var remoteGenericAttributesV3 = lo.Assign(
	remoteGenericAttributesV2,
	map[string]schema.Attribute{
		"propagate_query_params": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
)

var remoteGenericAttributesV4 = lo.Assign(
	remoteGenericAttributesV3,
	map[string]schema.Attribute{
		"retrieve_sha256_from_server": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo.",
		},
	},
)

func (r *remoteGenericResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     currentGenericSchemaVersion,
		Attributes:  remoteGenericAttributesV4,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}

func (r *remoteGenericResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 2 (prior state version) to 4 (Schema.Version)
		2: {
			PriorSchema: &schema.Schema{
				Attributes: remoteGenericAttributesV2,
				Blocks:     remoteBlocks,
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData RemoteGenericResourceModelV2

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := RemoteGenericResourceModelV4{
					RemoteGenericResourceModelV3: RemoteGenericResourceModelV3{
						RemoteGenericResourceModelV2: priorStateData,
						PropagateQueryParams:         types.BoolValue(false),
					},
					RetrieveSha256FromServer: types.BoolValue(false),
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
		// State upgrade implementation from 3 (prior state version) to 4 (Schema.Version)
		3: {
			PriorSchema: &schema.Schema{
				Attributes: remoteGenericAttributesV3,
				Blocks:     remoteBlocks,
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData RemoteGenericResourceModelV3

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := RemoteGenericResourceModelV4{
					RemoteGenericResourceModelV3: priorStateData,
					RetrieveSha256FromServer:     types.BoolValue(false),
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

// SDKv2
type GenericRemoteRepo struct {
	RepositoryRemoteBaseParams
	PropagateQueryParams     bool `json:"propagateQueryParams"`
	RetrieveSha256FromServer bool `hcl:"retrieve_sha256_from_server" json:"retrieveSha256FromServer"`
}

var genericSchemaV3 = lo.Assign(
	map[string]*sdkv2_schema.Schema{
		"propagate_query_params": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.GenericPackageType),
)

var GenericSchemaV4 = lo.Assign(
	genericSchemaV3,
	map[string]*sdkv2_schema.Schema{
		"retrieve_sha256_from_server": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo.",
		},
	},
)

const currentGenericSchemaVersion = 4

var GetGenericSchemas = func(s map[string]*sdkv2_schema.Schema) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			baseSchemaV1,
			genericSchemaV3,
		),
		1: lo.Assign(
			baseSchemaV1,
			genericSchemaV3,
		),
		2: lo.Assign(
			baseSchemaV2,
			genericSchemaV3,
		),
		3: lo.Assign(
			baseSchemaV3,
			genericSchemaV3,
		),
		4: lo.Assign(
			baseSchemaV3,
			s,
		),
	}
}

var GenericSchemas = GetGenericSchemas(GenericSchemaV4)

func GenericResourceStateUpgradeV3(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawState["retrieve_sha256_from_server"] = false
	if v, ok := rawState["property_sets"]; !ok || v == nil {
		rawState["property_sets"] = []string{}
	}

	return rawState, nil
}
