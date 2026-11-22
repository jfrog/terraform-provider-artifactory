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

package repository

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fw_path "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

func NewRepositoryExcludesResource() resource.Resource {
	return &RepositoryExcludesResource{
		TypeName: "artifactory_repository_excludes",
	}
}

type RepositoryExcludesResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type RepositoryExcludesResourceModel struct {
	RepoKey  types.String `tfsdk:"repo_key"`
	Patterns types.Set    `tfsdk:"patterns"`
}

type RepositoryExcludesAPIModel struct {
	Key             string `json:"key"`
	Rclass          string `json:"rclass"`
	ExcludesPattern string `json:"excludesPattern"`
}

func splitExcludesPattern(pattern string) []string {
	return lo.Compact(lo.Map(strings.Split(pattern, ","), func(p string, _ int) string {
		return strings.TrimSpace(p)
	}))
}

func (r *RepositoryExcludesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *RepositoryExcludesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repo_key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.RepoKey(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Repository key.",
			},
			"patterns": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
						validatorfw_string.RegexNotMatches(regexp.MustCompile(`,`), "must not contain a comma"),
					),
				},
				MarkdownDescription: "Set of Ant-style path patterns to add to the repository's `excludes_pattern`, e.g. `my-image/1.0.0/**` to block a Docker image tag. Patterns not managed by this resource are left untouched.",
			},
		},
		MarkdownDescription: "Manages a subset of the `excludes_pattern` of an existing repository, without managing the repository itself. " +
			"Useful to block specific packages or images (e.g. a Docker `image:tag`) declaratively.\n\n" +
			"~>Do not use this resource together with the `excludes_pattern` attribute on the repository resource for the same repository, as they will conflict. " +
			"Use `lifecycle { ignore_changes = [excludes_pattern] }` on the repository resource if it is also managed by Terraform.",
	}
}

func (r *RepositoryExcludesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *RepositoryExcludesResource) getRepo(repoKey string) (*RepositoryExcludesAPIModel, int, error) {
	var repo RepositoryExcludesAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("key", repoKey).
		SetResult(&repo).
		Get(RepositoriesEndpoint)
	if err != nil {
		return nil, 0, err
	}
	if response.IsError() {
		return nil, response.StatusCode(), errors.New(response.String())
	}
	return &repo, response.StatusCode(), nil
}

func (r *RepositoryExcludesResource) updatePatterns(repo *RepositoryExcludesAPIModel, add, remove []string) error {
	current := splitExcludesPattern(repo.ExcludesPattern)
	patterns := lo.Uniq(append(lo.Without(current, remove...), add...))

	body := RepositoryExcludesAPIModel{
		Key:             repo.Key,
		Rclass:          repo.Rclass,
		ExcludesPattern: strings.Join(patterns, ","),
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("key", repo.Key).
		SetBody(body).
		Post(RepositoriesEndpoint)
	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.New(response.String())
	}
	return nil
}

func patternsFromSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var patterns []string
	if set.IsNull() || set.IsUnknown() {
		return patterns, nil
	}
	diags := set.ElementsAs(ctx, &patterns, false)
	return patterns, diags
}

func (r *RepositoryExcludesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepositoryExcludesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patterns, diags := patternsFromSet(ctx, plan.Patterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, _, err := r.getRepo(plan.RepoKey.ValueString())
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if err := r.updatePatterns(repo, patterns, nil); err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryExcludesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepositoryExcludesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, statusCode, err := r.getRepo(state.RepoKey.ValueString())
	if statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	remote := splitExcludesPattern(repo.ExcludesPattern)

	statePatterns, diags := patternsFromSet(ctx, state.Patterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// On import there are no patterns in state yet, so adopt all existing patterns.
	patterns := remote
	if len(statePatterns) > 0 {
		patterns = lo.Intersect(statePatterns, remote)
	}

	patternsSet, diags := types.SetValueFrom(ctx, types.StringType, patterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Patterns = patternsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepositoryExcludesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan, state RepositoryExcludesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planPatterns, diags := patternsFromSet(ctx, plan.Patterns)
	resp.Diagnostics.Append(diags...)
	statePatterns, diags := patternsFromSet(ctx, state.Patterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, toRemove := lo.Difference(planPatterns, statePatterns)

	repo, _, err := r.getRepo(plan.RepoKey.ValueString())
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if err := r.updatePatterns(repo, planPatterns, toRemove); err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryExcludesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepositoryExcludesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statePatterns, diags := patternsFromSet(ctx, state.Patterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, statusCode, err := r.getRepo(state.RepoKey.ValueString())
	if statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if err := r.updatePatterns(repo, nil, statePatterns); err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}
}

// ImportState imports the resource into the Terraform state.
func (r *RepositoryExcludesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, fw_path.Root("repo_key"), req, resp)
}
