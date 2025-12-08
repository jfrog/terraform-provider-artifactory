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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

var SupportedGoVCSGitProviders = []string{
	"ARTIFACTORY",
	"BITBUCKET",
	"GITHUB",
	"GITHUBENTERPRISE",
	"GITLAB",
	"STASH",
}

func NewGoRemoteRepositoryResource() resource.Resource {
	return &remoteGoResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GoPackageType,
			repository.PackageNameLookup[repository.GoPackageType],
			reflect.TypeFor[remoteGoResourceModel](),
			reflect.TypeFor[RemoteGoAPIModel](),
		),
	}
}

type remoteGoResource struct {
	remoteResource
}

type remoteGoResourceModel struct {
	RemoteGenericResourceModelV4
	CurationResourceModel
	VCSGitProvider types.String `tfsdk:"vcs_git_provider"`
}

func (r *remoteGoResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteGoResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGoResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGoResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGoResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteGoResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGoResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteGoResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteGenericResourceModelV4.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteGoAPIModel{
		RemoteGenericAPIModel: remoteAPIModel.(RemoteGenericAPIModel),
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
		VCSGitProvider: r.VCSGitProvider.ValueString(),
	}, diags
}

func (r *remoteGoResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteGoAPIModel)

	r.RemoteGenericResourceModelV4.FromAPIModel(ctx, &model.RemoteGenericAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.VCSGitProvider = types.StringValue(model.VCSGitProvider)

	return diags
}

type RemoteGoAPIModel struct {
	RemoteGenericAPIModel
	CurationAPIModel
	VCSGitProvider string `json:"vcsGitProvider"`
}

func (r *remoteGoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteGoAttributes := lo.Assign(
		remoteGenericAttributesV4,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		map[string]schema.Attribute{
			"vcs_git_provider": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("ARTIFACTORY"),
				Validators: []validator.String{
					stringvalidator.OneOf(SupportedGoVCSGitProviders...),
				},
				MarkdownDescription: "Artifactory supports proxying the following Git providers out-of-the-box: GitHub (`GITHUB`), GitHub Enterprise (`GITHUBENTERPRISE`), BitBucket Cloud (`BITBUCKET`), BitBucket Server (`STASH`), GitLab (`GITLAB`), or a remote Artifactory instance (`ARTIFACTORY`). Default value is `ARTIFACTORY`.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteGoAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
