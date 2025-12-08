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

package lifecycle

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

const (
	ReleaseBundleV2PromotionEndpoint        = "/lifecycle/api/v2/promotion/records/{name}/{version}"
	ReleaseBundleV2PromotionDetailsEndpoint = "/lifecycle/api/v2/promotion/records/{name}/{version}/{created}"
)

var _ resource.Resource = &ReleaseBundleV2PromotionResource{}

func NewReleaseBundleV2PromotionResource() resource.Resource {
	return &ReleaseBundleV2PromotionResource{
		TypeName: "artifactory_release_bundle_v2_promotion",
	}
}

type ReleaseBundleV2PromotionResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type ReleaseBundleV2PromotionResourceModel struct {
	Name                   types.String `tfsdk:"name"`
	Version                types.String `tfsdk:"version"`
	KeyPairName            types.String `tfsdk:"keypair_name"`
	ProjectKey             types.String `tfsdk:"project_key"`
	Environment            types.String `tfsdk:"environment"`
	IncludedRepositoryKeys types.Set    `tfsdk:"included_repository_keys"`
	ExcludedRepositoryKeys types.Set    `tfsdk:"excluded_repository_keys"`
	Created                types.String `tfsdk:"created"`
	CreatedMillis          types.Int64  `tfsdk:"created_millis"`
}

func (m ReleaseBundleV2PromotionResourceModel) toAPIModel(ctx context.Context, apiModel *ReleaseBundleV2PromotionPostRequestAPIModel) (diags diag.Diagnostics) {
	var includedRepositoryKeys []string
	diags.Append(m.IncludedRepositoryKeys.ElementsAs(ctx, &includedRepositoryKeys, false)...)

	var excludedRepositoryKeys []string
	diags.Append(m.ExcludedRepositoryKeys.ElementsAs(ctx, &excludedRepositoryKeys, false)...)

	*apiModel = ReleaseBundleV2PromotionPostRequestAPIModel{
		Environment:            m.Environment.ValueString(),
		IncludedRepositoryKeys: includedRepositoryKeys,
		ExcludedRepositoryKeys: excludedRepositoryKeys,
	}

	return
}

type ReleaseBundleV2PromotionPostRequestAPIModel struct {
	Environment            string   `json:"environment"`
	IncludedRepositoryKeys []string `json:"included_repository_keys,omitempty"`
	ExcludedRepositoryKeys []string `json:"excluded_repository_keys,omitempty"`
}

type ReleaseBundleV2PromotionPostResponseAPIModel struct {
	Created       string `json:"created"`
	CreatedMillis int64  `json:"created_millis"`
}

type ReleaseBundleV2PromotionGetAPIModel struct {
	Environment   string `json:"environment"`
	Created       string `json:"created"`
	CreatedMillis int64  `json:"created_millis"`
}

func (r *ReleaseBundleV2PromotionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ReleaseBundleV2PromotionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Name of Release Bundle",
			},
			"version": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Version to promote",
			},
			"keypair_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Key-pair name to use for signature creation",
			},
			"project_key": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.ProjectKey(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Project key the Release Bundle belongs to",
			},
			"environment": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Target environment",
			},
			"included_repository_keys": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Defines specific repositories to include in the promotion. If this property is left undefined, all repositories (except those specifically excluded) are included in the promotion. Important: If one or more repositories are specifically included, all other repositories are excluded (regardless of what is defined in `excluded_repository_keys`).",
			},
			"excluded_repository_keys": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Defines specific repositories to exclude from the promotion.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the new version was created (ISO 8601 standard).",
			},
			"created_millis": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the new version was created (in milliseconds).",
			},
		},
		MarkdownDescription: "This resource enables you to promote Release Bundle V2 version. For more information, see [JFrog documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/promote-a-release-bundle-v2-to-a-target-environment).",
	}
}

func (r *ReleaseBundleV2PromotionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

var retryOnNotAssignedToEnvironmentError = func(response *resty.Response, _r error) bool {
	var notAssignedToEnvironmentRegex = regexp.MustCompile(".*not assigned to environment.*")

	return notAssignedToEnvironmentRegex.MatchString(string(response.Body()[:]))
}

func (r *ReleaseBundleV2PromotionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2PromotionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var promotion ReleaseBundleV2PromotionPostRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &promotion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.ProviderData.Client.R().
		SetHeader("X-JFrog-Signing-Key-Name", plan.KeyPairName.ValueString()).
		SetQueryParam("async", "false")

	if !plan.ProjectKey.IsNull() {
		request.SetQueryParam("project", plan.ProjectKey.ValueString())
	}

	var result ReleaseBundleV2PromotionPostResponseAPIModel

	response, err := request.
		SetPathParams(map[string]string{
			"name":    plan.Name.ValueString(),
			"version": plan.Version.ValueString(),
		}).
		SetBody(promotion).
		SetResult(&result).
		AddRetryCondition(retryOnNotAssignedToEnvironmentError).
		Post(ReleaseBundleV2PromotionEndpoint)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	plan.Created = types.StringValue(result.Created)
	plan.CreatedMillis = types.Int64Value(result.CreatedMillis)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleV2PromotionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2PromotionResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var promotion ReleaseBundleV2PromotionGetAPIModel

	request := r.ProviderData.Client.R()

	if !state.ProjectKey.IsNull() {
		request.SetQueryParam("project", state.ProjectKey.ValueString())
	}

	response, err := request.
		SetPathParams(map[string]string{
			"name":    state.Name.ValueString(),
			"version": state.Version.ValueString(),
			"created": fmt.Sprintf("%d", state.CreatedMillis.ValueInt64()),
		}).
		SetResult(&promotion).
		Get(ReleaseBundleV2PromotionDetailsEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	state.Environment = types.StringValue(promotion.Environment)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ReleaseBundleV2PromotionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning(
		"Update not supported",
		"Release Bundle V2 promotion cannnot be updated.",
	)
}

func (r *ReleaseBundleV2PromotionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2PromotionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	request := r.ProviderData.Client.R()

	if !state.ProjectKey.IsNull() {
		request.SetQueryParam("project", state.ProjectKey.ValueString())
	}

	response, err := request.
		SetPathParams(map[string]string{
			"name":    state.Name.ValueString(),
			"version": state.Version.ValueString(),
			"created": fmt.Sprintf("%d", state.CreatedMillis.ValueInt64()),
		}).
		SetQueryParam("async", "false").
		Delete(ReleaseBundleV2PromotionDetailsEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}
