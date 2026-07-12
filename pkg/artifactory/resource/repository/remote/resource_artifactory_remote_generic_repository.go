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
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

func NewGenericRemoteRepositoryResource() resource.Resource {
	return &remoteGenericResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GenericPackageType,
			repository.PackageNameLookup[repository.GenericPackageType],
			reflect.TypeFor[RemoteGenericResourceModelV5](),
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

type CustomHttpHeaderModel struct {
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	Sensitive types.Bool   `tfsdk:"sensitive"`
}

type RemoteGenericResourceModelV5 struct {
	RemoteGenericResourceModelV4
	CustomHttpHeaders []CustomHttpHeaderModel `tfsdk:"custom_http_headers"`
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

func (r *RemoteGenericResourceModelV5) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, r)...)
}

func (r *RemoteGenericResourceModelV5) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	base, d := r.RemoteGenericResourceModelV4.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}
	if diags.HasError() {
		return nil, diags
	}

	m := base.(RemoteGenericAPIModel)
	if len(r.CustomHttpHeaders) > 0 {
		headers := make([]httpHeaderAPIModel, 0, len(r.CustomHttpHeaders))
		for _, h := range r.CustomHttpHeaders {
			entry := httpHeaderAPIModel{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			}
			if h.Sensitive.ValueBool() {
				entry.Sensitive = true
			}
			headers = append(headers, entry)
		}
		m.CustomHttpHeaders = &headers
	}
	return m, diags
}

// FromAPIModel does not read custom_http_headers back from the API (write-only).
// r already holds the plan/state value before this method is called, so it is preserved as-is.
func (r *RemoteGenericResourceModelV5) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	model := apiModel.(*RemoteGenericAPIModel)
	return r.RemoteGenericResourceModelV4.FromAPIModel(ctx, model)
}

type httpHeaderAPIModel struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive,omitempty"`
}

const customHttpHeadersSupportedVersion = "7.146.0"

type customHttpHeadersVersionValidator struct {
	providerData *util.ProviderMetadata
}

func (v customHttpHeadersVersionValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Requires Artifactory %s or later to use custom_http_headers.", customHttpHeadersSupportedVersion)
}

func (v customHttpHeadersVersionValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Requires Artifactory `%s` or later to use `custom_http_headers`.", customHttpHeadersSupportedVersion)
}

func (v customHttpHeadersVersionValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	if v.providerData == nil {
		return
	}
	isSupported, err := util.CheckVersion(v.providerData.ArtifactoryVersion, customHttpHeadersSupportedVersion)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to check Artifactory version",
			fmt.Sprintf("Unable to validate custom_http_headers version requirement: %s", err.Error()),
		)
		return
	}
	if !isSupported {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unsupported Artifactory version",
			fmt.Sprintf("custom_http_headers requires Artifactory %s or later. Connected version: %s", customHttpHeadersSupportedVersion, v.providerData.ArtifactoryVersion),
		)
	}
}

type RemoteGenericAPIModel struct {
	RemoteAPIModel
	PropagateQueryParams     bool                  `json:"propagateQueryParams"`
	RetrieveSha256FromServer bool                  `json:"retrieveSha256FromServer"`
	CustomHttpHeaders        *[]httpHeaderAPIModel `json:"customHttpHeaders,omitempty"`
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
		Version: currentGenericSchemaVersion,
		Attributes: lo.Assign(remoteGenericAttributesV4, map[string]schema.Attribute{
			"custom_http_headers": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Up to 5 custom HTTP headers sent on every outbound request to the remote URL. Header values are write-only and masked in plan output. To remove all headers, remove this attribute. When `sensitive` is `true`, Artifactory encrypts the value server-side. Requires Artifactory %s or later.", customHttpHeadersSupportedVersion),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					customHttpHeadersVersionValidator{providerData: r.ProviderData},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Header name.",
						},
						"value": schema.StringAttribute{
							Required:            true,
							Sensitive:           true,
							MarkdownDescription: "Header value. Masked in Terraform plan output. Stored in state as configured; never read back from Artifactory.",
						},
						"sensitive": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "When `true`, Artifactory encrypts the header value server-side. Defaults to `false`.",
						},
					},
				},
			},
		}),
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}

func (r *remoteGenericResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
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

				upgradedStateData := RemoteGenericResourceModelV5{
					RemoteGenericResourceModelV4: RemoteGenericResourceModelV4{
						RemoteGenericResourceModelV3: RemoteGenericResourceModelV3{
							RemoteGenericResourceModelV2: priorStateData,
							PropagateQueryParams:         types.BoolValue(false),
						},
						RetrieveSha256FromServer: types.BoolValue(false),
					},
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
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

				upgradedStateData := RemoteGenericResourceModelV5{
					RemoteGenericResourceModelV4: RemoteGenericResourceModelV4{
						RemoteGenericResourceModelV3: priorStateData,
						RetrieveSha256FromServer:     types.BoolValue(false),
					},
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
		4: {
			PriorSchema: &schema.Schema{
				Attributes: remoteGenericAttributesV4,
				Blocks:     remoteBlocks,
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData RemoteGenericResourceModelV4

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := RemoteGenericResourceModelV5{
					RemoteGenericResourceModelV4: priorStateData,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

// SDKv2
type GenericRemoteRepo struct {
	RepositoryRemoteBaseParams
	PropagateQueryParams     bool                 `json:"propagateQueryParams"`
	RetrieveSha256FromServer bool                 `hcl:"retrieve_sha256_from_server" json:"retrieveSha256FromServer"`
	CustomHttpHeaders        []httpHeaderAPIModel `json:"customHttpHeaders,omitempty"`
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

var GenericSchemaV5 = lo.Assign(
	GenericSchemaV4,
	map[string]*sdkv2_schema.Schema{
		"custom_http_headers": {
			Type:      sdkv2_schema.TypeList,
			Optional:  true,
			Sensitive: true,
			MaxItems:  5,
			Elem: &sdkv2_schema.Resource{
				Schema: map[string]*sdkv2_schema.Schema{
					"name": {
						Type:     sdkv2_schema.TypeString,
						Required: true,
					},
					"value": {
						Type:      sdkv2_schema.TypeString,
						Required:  true,
						Sensitive: true,
					},
					"sensitive": {
						Type:     sdkv2_schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
			Description: "Up to 5 custom HTTP headers sent on every outbound request to the remote URL.",
		},
	},
)

const CurrentGenericRepositorySchemaVersion int16 = 5

const currentGenericSchemaVersion = int64(CurrentGenericRepositorySchemaVersion)

var GetGenericSchemas = func(s map[string]*sdkv2_schema.Schema) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(baseSchemaV1, genericSchemaV3),
		1: lo.Assign(baseSchemaV1, genericSchemaV3),
		2: lo.Assign(baseSchemaV2, genericSchemaV3),
		3: lo.Assign(baseSchemaV3, genericSchemaV3),
		4: lo.Assign(baseSchemaV3, GenericSchemaV4),
		5: lo.Assign(baseSchemaV3, s),
	}
}

var GenericSchemas = GetGenericSchemas(GenericSchemaV5)
