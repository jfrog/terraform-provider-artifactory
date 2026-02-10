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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

func NewNugetRemoteRepositoryResource() resource.Resource {
	return &remoteNugetResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.NugetPackageType,
			repository.PackageNameLookup[repository.NugetPackageType],
			reflect.TypeFor[remoteNugetResourceModel](),
			reflect.TypeFor[RemoteNugetAPIModel](),
		),
	}
}

type remoteNugetResource struct {
	remoteResource
}

type remoteNugetResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	FeedContextPath          types.String `tfsdk:"feed_context_path"`
	DownloadContextPath      types.String `tfsdk:"download_context_path"`
	V3FeedURL                types.String `tfsdk:"v3_feed_url"`
	ForceNugetAuthentication types.Bool   `tfsdk:"force_nuget_authentication"`
	SymbolServerURL          types.String `tfsdk:"symbol_server_url"`
}

func (r *remoteNugetResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteNugetResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteNugetResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteNugetResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteNugetResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteNugetResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteNugetResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteNugetResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteNugetAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated:     r.Curated.ValueBool(),
			PassThrough: r.PassThrough.ValueBool(),
		},
		FeedContextPath:          r.FeedContextPath.ValueString(),
		DownloadContextPath:      r.DownloadContextPath.ValueString(),
		V3FeedURL:                r.V3FeedURL.ValueString(),
		ForceNugetAuthentication: r.ForceNugetAuthentication.ValueBool(),
		SymbolServerURL:          r.SymbolServerURL.ValueString(),
	}, diags
}

func (r *remoteNugetResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteNugetAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.PassThrough = types.BoolValue(model.CurationAPIModel.PassThrough)
	r.FeedContextPath = types.StringValue(model.FeedContextPath)
	r.DownloadContextPath = types.StringValue(model.DownloadContextPath)
	r.V3FeedURL = types.StringValue(model.V3FeedURL)
	r.ForceNugetAuthentication = types.BoolValue(model.ForceNugetAuthentication)
	r.SymbolServerURL = types.StringValue(model.SymbolServerURL)
	return diags
}

type RemoteNugetAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	FeedContextPath          string `json:"feedContextPath"`
	DownloadContextPath      string `json:"downloadContextPath"`
	V3FeedURL                string `json:"v3FeedUrl"`
	ForceNugetAuthentication bool   `json:"forceNugetAuthentication"`
	SymbolServerURL          string `json:"symbolServerUrl"`
}

func (r *remoteNugetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteNugetAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"feed_context_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("api/v2"),
				MarkdownDescription: "When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is 'api/v2'.",
			},
			"download_context_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("api/v2/package"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "The context path prefix through which NuGet downloads are served. Default value is 'api/v2/package'.",
			},
			"v3_feed_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://api.nuget.org/v3/index.json"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "The URL to the NuGet v3 feed. Default value is 'https://api.nuget.org/v3/index.json'.",
			},
			"force_nuget_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force basic authentication credentials in order to use this repository. Default value is 'false'",
			},
			"symbol_server_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://symbols.nuget.org/download/symbols"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "NuGet symbol server URL.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteNugetAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
