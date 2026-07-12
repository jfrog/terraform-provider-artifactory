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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewComposerRemoteRepositoryResource() resource.Resource {
	return &remoteComposerResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.ComposerPackageType,
			repository.PackageNameLookup[repository.ComposerPackageType],
			reflect.TypeFor[remoteComposerResourceModel](),
			reflect.TypeFor[RemoteComposerAPIModel](),
		),
	}
}

type remoteComposerResource struct {
	remoteResource
}

type remoteComposerResourceModel struct {
	RemoteResourceModel
	vcsResourceModel
	ComposerRegistryUrl types.String `tfsdk:"composer_registry_url"`
}

func (r *remoteComposerResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteComposerResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteComposerResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteComposerResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteComposerResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteComposerResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteComposerResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteComposerResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteComposerAPIModel{
		RemoteAPIModel: remoteAPIModel,
		vcsAPIModel: vcsAPIModel{
			GitProvider:    r.VCSGitProvider.ValueStringPointer(),
			GitDownloadURL: r.VCSGitDownloadURL.ValueStringPointer(),
		},
		ComposerRegistryUrl: r.ComposerRegistryUrl.ValueString(),
	}, diags
}

func (r *remoteComposerResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteComposerAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.VCSGitProvider = types.StringPointerValue(model.vcsAPIModel.GitProvider)
	r.VCSGitDownloadURL = types.StringPointerValue(model.vcsAPIModel.GitDownloadURL)
	r.ComposerRegistryUrl = types.StringValue(model.ComposerRegistryUrl)

	return diags
}

type RemoteComposerAPIModel struct {
	RemoteAPIModel
	vcsAPIModel
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

func (r *remoteComposerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteComposerAttributes := lo.Assign(
		RemoteAttributes,
		vcsAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"composer_registry_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://packagist.org"),
				MarkdownDescription: "Proxy remote Composer repository. Default value is 'https://packagist.org'.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteComposerAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
