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

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewOCILocalRepositoryResource() resource.Resource {
	return &localOCIResource{
		localResource: NewLocalRepositoryResource(
			repository.OCIPackageType,
			"OCI",
			reflect.TypeFor[LocalOCIResourceModel](),
			reflect.TypeFor[LocalOCIAPIModel](),
		),
	}
}

type localOCIResource struct {
	localResource
}

type LocalOCIResourceModel struct {
	LocalResourceModel
	MaxUniqueTags types.Int64 `tfsdk:"max_unique_tags"`
	TagRetention  types.Int64 `tfsdk:"tag_retention"`
}

func (r *LocalOCIResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalOCIResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalOCIResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalOCIResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalOCIResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalOCIResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalOCIResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalOCIResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalOCIAPIModel{
		LocalAPIModel:    localAPIModel,
		MaxUniqueTags:    r.MaxUniqueTags.ValueInt64(),
		DockerApiVersion: "V2",
		TagRetention:     r.TagRetention.ValueInt64(),
	}, diags
}

func (r *LocalOCIResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalOCIAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.MaxUniqueTags = types.Int64Value(model.MaxUniqueTags)
	r.TagRetention = types.Int64Value(model.TagRetention)

	return diags
}

type LocalOCIAPIModel struct {
	LocalAPIModel
	MaxUniqueTags    int64  `json:"maxUniqueTags"`
	DockerApiVersion string `json:"dockerApiVersion"`
	TagRetention     int64  `json:"dockerTagRetention"`
}

func (r *localOCIResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"max_unique_tags": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The maximum number of unique tags of a single OCI image to store in this repository. Once the number tags for an object exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.",
			},
			"tag_retention": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				MarkdownDescription: "If greater than 1, overwritten tags will be saved by their digest, up to the set up number.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

type OciLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags    int    `json:"maxUniqueTags"`
	DockerApiVersion string `json:"dockerApiVersion"`
	TagRetention     int    `json:"dockerTagRetention"`
}

var ociSchema = lo.Assign(
	map[string]*sdkv2_schema.Schema{
		"max_unique_tags": {
			Type:     sdkv2_schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique tags of a single OCI image to store in this repository.\n" +
				"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"tag_retention": {
			Type:             sdkv2_schema.TypeInt,
			Optional:         true,
			Computed:         false,
			Description:      "If greater than 1, overwritten tags will be saved by their digest, up to the set up number.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.OCIPackageType),
)

var OCILocalSchemas = GetSchemas(ociSchema)
