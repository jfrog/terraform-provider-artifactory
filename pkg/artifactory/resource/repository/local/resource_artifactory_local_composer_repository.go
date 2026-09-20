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

package local

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

func NewComposerLocalRepositoryResource() resource.Resource {
	return &localComposerResource{
		localResource: NewLocalRepositoryResource(
			repository.ComposerPackageType,
			repository.PackageNameLookup[repository.ComposerPackageType],
			reflect.TypeFor[LocalComposerResourceModel](),
			reflect.TypeFor[LocalComposerAPIModel](),
		),
	}
}

type localComposerResource struct {
	localResource
}

type LocalComposerResourceModel struct {
	LocalResourceModel
	EnableComposerV1Indexing types.Bool `tfsdk:"enable_composer_v1_indexing"`
}

func (r *LocalComposerResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalComposerResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalComposerResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalComposerResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalComposerResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalComposerResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalComposerResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalComposerResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalComposerAPIModel{
		LocalAPIModel:            localAPIModel,
		EnableComposerV1Indexing: r.EnableComposerV1Indexing.ValueBool(),
	}, diags
}

func (r *LocalComposerResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalComposerAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.EnableComposerV1Indexing = types.BoolValue(model.EnableComposerV1Indexing)

	return diags
}

type LocalComposerAPIModel struct {
	LocalAPIModel
	EnableComposerV1Indexing bool `json:"enableComposerV1Indexing"`
}

func (r *localComposerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"enable_composer_v1_indexing": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable Composer metadata version 1 indexing. Default value is 'false'.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var composerSchema = map[string]*sdkv2_schema.Schema{
	"enable_composer_v1_indexing": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable Composer metadata version 1 indexing. Default value is 'false'.",
	},
}

var ComposerSchemas = GetSchemas(lo.Assign(
	composerSchema,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.ComposerPackageType),
))

type ComposerLocalRepositoryParams struct {
	RepositoryBaseParams
	EnableComposerV1Indexing bool `hcl:"enable_composer_v1_indexing" json:"enableComposerV1Indexing"`
}
