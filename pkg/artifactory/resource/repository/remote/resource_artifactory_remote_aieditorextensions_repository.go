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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

// AIEditorExtensionsDefaultExternalDependenciesPattern is the pattern Artifactory
// applies by default for this package type. Extension payloads are served from
// the `vsassets.io` CDN rather than the gallery host itself, so external
// dependency resolution is required for downloads to succeed out of the box.
const AIEditorExtensionsDefaultExternalDependenciesPattern = "**/**vsassets.io/**"

// AIEditorExtensionsMarketplaceURL is the VS Code marketplace gallery endpoint.
// Artifactory applies no default of its own for this package type and rejects a
// create without a URL, so `url` stays required and this value is only a
// documented recommendation.
const AIEditorExtensionsMarketplaceURL = "https://marketplace.visualstudio.com/_apis/public/gallery"

var aiEditorExtensionsDefaultExternalDependenciesPatterns = types.ListValueMust(
	types.StringType,
	[]attr.Value{types.StringValue(AIEditorExtensionsDefaultExternalDependenciesPattern)},
)

func NewAIEditorExtensionsRemoteRepositoryResource() resource.Resource {
	return &remoteAIEditorExtensionsResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.AIEditorExtensionsPackageType,
			repository.PackageNameLookup[repository.AIEditorExtensionsPackageType],
			reflect.TypeFor[RemoteAIEditorExtensionsResourceModel](),
			reflect.TypeFor[RemoteAIEditorExtensionsAPIModel](),
		),
	}
}

type remoteAIEditorExtensionsResource struct {
	remoteResource
}

type RemoteAIEditorExtensionsResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	ExternalDependenciesEnabled  types.Bool              `tfsdk:"external_dependencies_enabled"`
	ExternalDependenciesPatterns types.List              `tfsdk:"external_dependencies_patterns"`
	PropagateQueryParams         types.Bool              `tfsdk:"propagate_query_params"`
	RetrieveSha256FromServer     types.Bool              `tfsdk:"retrieve_sha256_from_server"`
	CustomHttpHeaders            []CustomHttpHeaderModel `tfsdk:"custom_http_headers"`
}

func (r *RemoteAIEditorExtensionsResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r RemoteAIEditorExtensionsResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteAIEditorExtensionsResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteAIEditorExtensionsResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteAIEditorExtensionsResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteAIEditorExtensionsResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteAIEditorExtensionsResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r RemoteAIEditorExtensionsResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	var externalDependenciesPatterns []string
	d = r.ExternalDependenciesPatterns.ElementsAs(ctx, &externalDependenciesPatterns, false)
	if d != nil {
		diags.Append(d...)
	}

	// The patterns are sent unconditionally, including when external dependencies
	// are disabled. Artifactory stores them either way and keeps whatever it
	// already has when an update omits them, so skipping the field would leave
	// the server on stale custom patterns while state moved to the computed
	// default — a diff that never converges.
	apiModel := RemoteAIEditorExtensionsAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated:     r.Curated.ValueBool(),
			PassThrough: r.PassThrough.ValueBool(),
		},
		ExternalDependenciesEnabled:  r.ExternalDependenciesEnabled.ValueBool(),
		ExternalDependenciesPatterns: externalDependenciesPatterns,
		PropagateQueryParams:         r.PropagateQueryParams.ValueBool(),
		RetrieveSha256FromServer:     r.RetrieveSha256FromServer.ValueBool(),
	}

	if len(r.CustomHttpHeaders) > 0 {
		headers := make([]httpHeaderAPIModel, 0, len(r.CustomHttpHeaders))
		for _, h := range r.CustomHttpHeaders {
			headers = append(headers, httpHeaderAPIModel{
				Name:      h.Name.ValueString(),
				Value:     h.Value.ValueString(),
				Sensitive: h.Sensitive.ValueBool(),
			})
		}
		apiModel.CustomHttpHeaders = &headers
	}

	return apiModel, diags
}

func (r *RemoteAIEditorExtensionsResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteAIEditorExtensionsAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.ExternalDependenciesEnabled = types.BoolValue(model.ExternalDependenciesEnabled)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.PassThrough = types.BoolValue(model.CurationAPIModel.PassThrough)
	r.PropagateQueryParams = types.BoolValue(model.PropagateQueryParams)
	r.RetrieveSha256FromServer = types.BoolValue(model.RetrieveSha256FromServer)

	// `custom_http_headers` is deliberately not read back. Artifactory returns the
	// values in plaintext, so refreshing from the API would write header secrets
	// into state from a source other than the configuration. `r` already carries
	// the plan/state value when this runs, so leaving it untouched preserves it.

	// Artifactory always returns the patterns, disabled or not, so the fallback
	// only guards against a repository stored without them at all — it keeps the
	// value non-null, as the computed schema default requires.
	r.ExternalDependenciesPatterns = aiEditorExtensionsDefaultExternalDependenciesPatterns
	if len(model.ExternalDependenciesPatterns) > 0 {
		externalDependenciesPatterns, d := types.ListValueFrom(ctx, types.StringType, model.ExternalDependenciesPatterns)
		if d != nil {
			diags.Append(d...)
		}
		r.ExternalDependenciesPatterns = externalDependenciesPatterns
	}

	return diags
}

type RemoteAIEditorExtensionsAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	ExternalDependenciesEnabled  bool                  `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string              `json:"externalDependenciesPatterns,omitempty"`
	PropagateQueryParams         bool                  `json:"propagateQueryParams"`
	RetrieveSha256FromServer     bool                  `json:"retrieveSha256FromServer"`
	CustomHttpHeaders            *[]httpHeaderAPIModel `json:"customHttpHeaders,omitempty"`
}

func (r *remoteAIEditorExtensionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteAIEditorExtensionsAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "The URL of the marketplace gallery to proxy. For the VS Code marketplace use " +
					"`" + AIEditorExtensionsMarketplaceURL + "`. This attribute is required: Artifactory applies no " +
					"default URL for this package type and rejects the request with `No URL defined for remote repository` " +
					"if it is omitted.",
			},
			// Artifactory pins bypassHeadRequests to true for this package type and rejects any attempt
			// to change it, so it is exposed as read-only (Computed-only) and always reports `true`.
			// It cannot be set in configuration.
			"bypass_head_requests": schema.BoolAttribute{
				Computed: true,
				Default:  booldefault.StaticBool(true),
				MarkdownDescription: "Artifactory always enables this setting for AI-Editor Extensions repositories and " +
					"rejects any attempt to change it, so it is **read-only and always `true`** (it cannot be set in configuration).",
			},
			"external_dependencies_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When set, Artifactory can resolve extension dependencies from the external sources matching `external_dependencies_patterns`. Unlike other package types, this defaults to `true` because extension payloads are hosted on a separate CDN from the gallery URL.",
			},
			"external_dependencies_patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(aiEditorExtensionsDefaultExternalDependenciesPatterns),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "An allow list of Ant-style path patterns that determine which remote hosts external extension " +
					"dependencies may be downloaded from. Only applies when `external_dependencies_enabled` is `true`. " +
					"Default value is `[\"" + AIEditorExtensionsDefaultExternalDependenciesPattern + "\"]`, which covers the CDN serving VS Code marketplace extension payloads.",
			},
			"propagate_query_params": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
			},
			"retrieve_sha256_from_server": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo.",
			},
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
							MarkdownDescription: "Header name. Artifactory stores header names lower-cased.",
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
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteAIEditorExtensionsAttributes,
		Blocks:      remoteBlocks,
		Description: "Provides a remote AI-Editor Extensions repository that proxies and caches editor extensions from a VS Code compatible marketplace gallery, such as `https://marketplace.visualstudio.com/_apis/public/gallery`. This package type is supported as a remote repository only, so the provider exposes no local, virtual, or federated equivalent.",
	}
}

// There is deliberately no ValidateConfig rejecting patterns while
// `external_dependencies_enabled` is false. Other package types do reject that
// combination, but Artifactory accepts and stores it, and now that the patterns
// are always sent it round-trips cleanly. Forbidding it would stop Terraform from
// expressing a state the API supports, for no benefit.
