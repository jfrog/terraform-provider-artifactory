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
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkv2_validation "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

const MavenCurrentSchemaVersion = 2

func NewMavenRemoteRepositoryResource() resource.Resource {
	return &remoteMavenResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.MavenPackageType,
			repository.PackageNameLookup[repository.MavenPackageType],
			reflect.TypeFor[remoteMavenResourceModel](),
			reflect.TypeFor[RemoteMavenAPIModel](),
		),
	}
}

type remoteMavenResource struct {
	remoteResource
}

type remoteMavenResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	JavaResourceModel
}

func (r *remoteMavenResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteMavenResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteMavenResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteMavenResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteMavenResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteMavenResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteMavenResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteMavenResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteMavenAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
		JavaAPIModel: JavaAPIModel{
			FetchJarsEagerly:             r.FetchJarsEagerly.ValueBool(),
			FetchSourcesEagerly:          r.FetchSourcesEagerly.ValueBool(),
			RemoteRepoChecksumPolicyType: r.RemoteRepoChecksumPolicyType.ValueString(),
			HandleReleases:               r.HandleReleases.ValueBool(),
			HandleSnapshots:              r.HandleSnapshots.ValueBool(),
			SuppressPomConsistencyChecks: r.SuppressPomConsistencyChecks.ValueBool(),
			RejectInvalidJars:            r.RejectInvalidJars.ValueBool(),
			MaxUniqueSnapshots:           r.MaxUniqueSnapshots.ValueInt64(),
		},
	}, diags
}

func (r *remoteMavenResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteMavenAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.FetchJarsEagerly = types.BoolValue(model.JavaAPIModel.FetchJarsEagerly)
	r.FetchSourcesEagerly = types.BoolValue(model.JavaAPIModel.FetchSourcesEagerly)
	r.RemoteRepoChecksumPolicyType = types.StringValue(model.JavaAPIModel.RemoteRepoChecksumPolicyType)
	r.HandleReleases = types.BoolValue(model.JavaAPIModel.HandleReleases)
	r.HandleSnapshots = types.BoolValue(model.JavaAPIModel.HandleSnapshots)
	r.SuppressPomConsistencyChecks = types.BoolValue(model.JavaAPIModel.SuppressPomConsistencyChecks)
	r.RejectInvalidJars = types.BoolValue(model.JavaAPIModel.RejectInvalidJars)
	r.MaxUniqueSnapshots = types.Int64Value(model.JavaAPIModel.MaxUniqueSnapshots)

	return diags
}

type RemoteMavenAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	JavaAPIModel
}

func (r *remoteMavenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteMavenAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		javaAttributes(false),
	)

	resp.Schema = schema.Schema{
		Version:     MavenCurrentSchemaVersion,
		Attributes:  remoteMavenAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}

// SDKv2

// Old schema, the one needs to be migrated (seconds -> secs)
var mavenSchemaV1 = lo.Assign(
	JavaSchema(repository.MavenPackageType, false),
	map[string]*sdkv2_schema.Schema{
		"metadata_retrieval_timeout_seconds": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      60,
			ValidateFunc: sdkv2_validation.IntAtLeast(0),
			Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on remote server. A value of 0 indicates no caching. Cannot be larger than retrieval_cache_period_seconds attribute. Default value is 60.",
		},
	},
)

var mavenSchemaV2 = lo.Assign(
	JavaSchema(repository.MavenPackageType, false),
	CurationRemoteRepoSchema,
)

var MavenSchemas = map[int16]map[string]*sdkv2_schema.Schema{
	0: lo.Assign(
		baseSchemaV1,
		mavenSchemaV1,
	),
	1: lo.Assign(
		baseSchemaV1,
		mavenSchemaV1,
	),
	2: lo.Assign(
		baseSchemaV2,
		mavenSchemaV2,
	),
}
