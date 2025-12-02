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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	sdkv2_validator "github.com/jfrog/terraform-provider-shared/validator"
	"github.com/samber/lo"
)

func NewRPMLocalRepositoryResource() resource.Resource {
	return &localRPMResource{
		localResource: NewLocalRepositoryResource(
			repository.RPMPackageType,
			"RPM",
			reflect.TypeFor[LocalRPMResourceModel](),
			reflect.TypeFor[LocalRPMAPIModel](),
		),
	}
}

type localRPMResource struct {
	localResource
}

type LocalRPMResourceModel struct {
	LocalResourceModel
	PrimaryKeyPairRef       types.String `tfsdk:"primary_keypair_ref"`
	SecondaryKeyPairRef     types.String `tfsdk:"secondary_keypair_ref"`
	RootDepth               types.Int64  `tfsdk:"yum_root_depth"`
	CalculateYumMetadata    types.Bool   `tfsdk:"calculate_yum_metadata"`
	EnableFileListsIndexing types.Bool   `tfsdk:"enable_file_lists_indexing"`
	GroupFileNames          types.String `tfsdk:"yum_group_file_names"`
}

func (r *LocalRPMResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalRPMResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalRPMResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalRPMResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalRPMResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalRPMResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalRPMResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalRPMResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalRPMAPIModel{
		LocalAPIModel:           localAPIModel,
		PrimaryKeyPairRef:       r.PrimaryKeyPairRef.ValueString(),
		SecondaryKeyPairRef:     r.SecondaryKeyPairRef.ValueString(),
		RootDepth:               r.RootDepth.ValueInt64(),
		CalculateYumMetadata:    r.CalculateYumMetadata.ValueBool(),
		EnableFileListsIndexing: r.EnableFileListsIndexing.ValueBool(),
		GroupFileNames:          r.GroupFileNames.ValueString(),
	}, diags
}

func (r *LocalRPMResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalRPMAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.PrimaryKeyPairRef = types.StringValue(model.PrimaryKeyPairRef)
	r.SecondaryKeyPairRef = types.StringValue(model.SecondaryKeyPairRef)
	r.RootDepth = types.Int64Value(model.RootDepth)
	r.CalculateYumMetadata = types.BoolValue(model.CalculateYumMetadata)
	r.EnableFileListsIndexing = types.BoolValue(model.EnableFileListsIndexing)
	r.GroupFileNames = types.StringValue(model.GroupFileNames)

	return diags
}

type LocalRPMAPIModel struct {
	LocalAPIModel
	PrimaryKeyPairRef       string `json:"primaryKeyPairRef"`
	SecondaryKeyPairRef     string `json:"secondaryKeyPairRef"`
	RootDepth               int64  `json:"yumRootDepth"`
	CalculateYumMetadata    bool   `json:"calculateYumMetadata"`
	EnableFileListsIndexing bool   `json:"enableFileListsIndexing"`
	GroupFileNames          string `json:"yumGroupFileNames"`
}

func (r *localRPMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		repository.PrimaryKeyPairRefAttribute,
		repository.SecondaryKeyPairRefAttribute,
		map[string]schema.Attribute{
			"yum_root_depth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The depth, relative to the repository's root folder, where RPM metadata is created. " +
					"This is useful when your repository contains multiple RPM repositories under parallel hierarchies. " +
					"For example, if your RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4.",
			},
			"calculate_yum_metadata": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"enable_file_lists_indexing": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"yum_group_file_names": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.+(?:,.+)*`), "must be comma separated string"),
				},
				MarkdownDescription: "A comma separated list of XML file names containing RPM group component definitions. Artifactory includes " +
					"the group definitions as part of the calculated RPM metadata, as well as automatically generating a " +
					"gzipped version of the group files, if required.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

var rpmSchema = utilsdk.MergeMaps(
	repository.PrimaryKeyPairRefSDKv2,
	repository.SecondaryKeyPairRefSDKv2,
	map[string]*sdkv2_schema.Schema{
		"yum_root_depth": {
			Type:             sdkv2_schema.TypeInt,
			Optional:         true,
			Default:          0,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			Description: "The depth, relative to the repository's root folder, where RPM metadata is created. " +
				"This is useful when your repository contains multiple RPM repositories under parallel hierarchies. " +
				"For example, if your RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4.",
		},
		"calculate_yum_metadata": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"enable_file_lists_indexing": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"yum_group_file_names": {
			Type:             sdkv2_schema.TypeString,
			Optional:         true,
			Default:          "",
			ValidateDiagFunc: sdkv2_validator.CommaSeperatedList,
			Description: "A comma separated list of XML file names containing RPM group component definitions. Artifactory includes " +
				"the group definitions as part of the calculated RPM metadata, as well as automatically generating a " +
				"gzipped version of the group files, if required.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.RPMPackageType),
)

var RPMSchemas = GetSchemas(rpmSchema)

type RpmLocalRepositoryParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
	repository.SecondaryKeyPairRefParam
	RootDepth               int    `hcl:"yum_root_depth" json:"yumRootDepth"`
	CalculateYumMetadata    bool   `hcl:"calculate_yum_metadata" json:"calculateYumMetadata"`
	EnableFileListsIndexing bool   `hcl:"enable_file_lists_indexing" json:"enableFileListsIndexing"`
	GroupFileNames          string `hcl:"yum_group_file_names" json:"yumGroupFileNames"`
}
