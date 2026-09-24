// Copyright (c) JFrog Ltd. (2026)
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

// NimModelDefaultRegistryURL is NVIDIA's public NGC catalog API, the registry
// Artifactory suggests when creating a NimModel repository. Unlike types with a
// server-enforced default URL, Artifactory applies none of its own here and
// rejects a create without a URL, so `url` stays required and this value is only
// a documented recommendation.
const NimModelDefaultRegistryURL = "https://api.ngc.nvidia.com"

func NewNimModelRemoteRepositoryResource() resource.Resource {
	return &remoteNimModelResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.NimModelPackageType,
			repository.PackageNameLookup[repository.NimModelPackageType],
			reflect.TypeFor[RemoteNimModelResourceModel](),
			reflect.TypeFor[RemoteNimModelAPIModel](),
		),
	}
}

type remoteNimModelResource struct {
	remoteResource
}

type RemoteNimModelResourceModel struct {
	RemoteResourceModel
	EnableTokenAuthentication types.Bool `tfsdk:"enable_token_authentication"`
}

func (r *RemoteNimModelResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r RemoteNimModelResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteNimModelResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteNimModelResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteNimModelResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteNimModelResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteNimModelResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r RemoteNimModelResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	apiModel := RemoteNimModelAPIModel{
		RemoteAPIModel:            remoteAPIModel,
		EnableTokenAuthentication: r.EnableTokenAuthentication.ValueBool(),
	}

	return apiModel, diags
}

func (r *RemoteNimModelResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteNimModelAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.EnableTokenAuthentication = types.BoolValue(model.EnableTokenAuthentication)

	return diags
}

type RemoteNimModelAPIModel struct {
	RemoteAPIModel
	EnableTokenAuthentication bool `json:"enableTokenAuthentication"`
}

func (r *remoteNimModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteNimModelAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "The URL of the NVIDIA NIM registry to proxy, e.g. `" + NimModelDefaultRegistryURL + "` for the public NVIDIA NGC catalog. " +
					"This attribute is required: Artifactory applies no default URL for this package type and rejects the request with `No URL defined for remote repository` " +
					"if it is omitted.",
			},
			// Artifactory defaults this to `true` for NimModel repositories (unlike most remote
			// package types, which default it to `false`), matching the Docker/OCI/HelmOCI pattern.
			"enable_token_authentication": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				MarkdownDescription: "Enable token (Bearer) based authentication. Defaults to `true`, matching the " +
					"Artifactory default for this package type.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteNimModelAttributes,
		Blocks:      remoteBlocks,
		Description: "Provides a remote NVIDIA NIM repository that proxies and caches NIM models from an NGC-compatible model registry, such as `" + NimModelDefaultRegistryURL + "`. This package type is supported as a remote repository only, so the provider exposes no local, virtual, or federated equivalent.",
	}
}
