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

// BazelResourceName is the Terraform resource type name. It is intentionally
// shorter than the Artifactory package type (`bazelmodules`); the package type
// sent to the API remains `repository.BazelModulesPackageType`.
const BazelResourceName = "artifactory_remote_bazel_repository"

func NewBazelRemoteRepositoryResource() resource.Resource {
	r := &remoteBazelResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.BazelModulesPackageType,
			repository.PackageNameLookup[repository.BazelModulesPackageType],
			reflect.TypeFor[RemoteBazelResourceModel](),
			reflect.TypeFor[RemoteBazelAPIModel](),
		),
	}
	// The Artifactory package type is `bazelmodules`, but the resource is exposed
	// with the shorter `artifactory_remote_bazel_repository` name.
	r.TypeName = BazelResourceName
	return r
}

type remoteBazelResource struct {
	remoteResource
}

type RemoteBazelResourceModel struct {
	RemoteResourceModel
}

func (r *RemoteBazelResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r RemoteBazelResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteBazelResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteBazelResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteBazelResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteBazelResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteBazelResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r RemoteBazelResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteBazelAPIModel{
		RemoteAPIModel: remoteAPIModel,
	}, diags
}

func (r *RemoteBazelResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteBazelAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)

	return diags
}

type RemoteBazelAPIModel struct {
	RemoteAPIModel
}

func (r *remoteBazelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteBazelAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteBazelAttributes,
		Blocks:      remoteBlocks,
		Description: "Provides a remote Bazel Modules repository that proxies and caches modules from a Bazel registry such as the Bazel Central Registry (`https://bcr.bazel.build/`). Bazel Modules repositories are supported as remote repositories only.",
	}
}
