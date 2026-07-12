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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewCargoRemoteRepositoryResource() resource.Resource {
	return &remoteCargoResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.CargoPackageType,
			repository.PackageNameLookup[repository.CargoPackageType],
			reflect.TypeFor[remoteCargoResourceModel](),
			reflect.TypeFor[RemoteCargoAPIModel](),
		),
	}
}

type remoteCargoResource struct {
	remoteResource
}

type remoteCargoResourceModel struct {
	RemoteResourceModel
	GitRegistryURL    types.String `tfsdk:"git_registry_url"`
	AnonymousAccess   types.Bool   `tfsdk:"anonymous_access"`
	EnableSparseIndex types.Bool   `tfsdk:"enable_sparse_index"`
}

func (r *remoteCargoResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteCargoResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCargoResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCargoResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCargoResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteCargoResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCargoResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteCargoResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteCargoAPIModel{
		RemoteAPIModel:    remoteAPIModel,
		GitRegistryURL:    r.GitRegistryURL.ValueString(),
		AnonymousAccess:   r.AnonymousAccess.ValueBool(),
		EnableSparseIndex: r.EnableSparseIndex.ValueBool(),
	}, diags
}

func (r *remoteCargoResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteCargoAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.GitRegistryURL = types.StringValue(model.GitRegistryURL)
	r.AnonymousAccess = types.BoolValue(model.AnonymousAccess)
	r.EnableSparseIndex = types.BoolValue(model.EnableSparseIndex)

	return diags
}

type RemoteCargoAPIModel struct {
	RemoteAPIModel
	GitRegistryURL    string `json:"gitRegistryUrl"`
	AnonymousAccess   bool   `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool   `json:"cargoInternalIndex"`
}

func (r *remoteCargoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteCargoAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"git_registry_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://registry.Cargo.io"),
				MarkdownDescription: "This is the index url, expected to be a git repository. Default value in UI is 'https://index.crates.io/'",
			},
			"anonymous_access": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "(On the UI: Anonymous download and search) Cargo client does not send credentials when performing download and search for crates. " +
					"Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.",
			},
			"enable_sparse_index": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is `false`.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteCargoAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
