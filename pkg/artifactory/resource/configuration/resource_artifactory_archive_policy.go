package configuration

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

var yumPolicyPackageType = "yum" // Only used by cleanup and archive policies as RPM

var archivePolicySupportedPackageType = []string{
	repository.CargoPackageType,
	repository.CocoapodsPackageType,
	repository.ConanPackageType,
	repository.DebianPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HelmPackageType,
	repository.HelmOCIPackageType,
	repository.HuggingFacePackageType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.OCIPackageType,
	repository.PyPiPackageType,
	repository.TerraformPackageType,
	yumPolicyPackageType,
}

func NewArchivePolicyResource() resource.Resource {
	return &ArchivePolicyResource{
		JFrogResource: util.JFrogResource{
			TypeName:                "artifactory_archive_policy",
			ValidArtifactoryVersion: "7.102.0",
			DocumentEndpoint:        "artifactory/api/archive/v2/packages/policies/{policyKey}",
		},
		EnablementEndpoint: "artifactory/api/archive/v2/packages/policies/{policyKey}/enablement",
	}
}

type ArchivePolicyResource struct {
	util.JFrogResource
	EnablementEndpoint string
}

type ArchivePolicyResourceModel struct {
	Key               types.String `tfsdk:"key"`
	Description       types.String `tfsdk:"description"`
	CronExpression    types.String `tfsdk:"cron_expression"`
	DurationInMinutes types.Int64  `tfsdk:"duration_in_minutes"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SkipTrashcan      types.Bool   `tfsdk:"skip_trashcan"`
	ProjectKey        types.String `tfsdk:"project_key"`
	SearchCriteria    types.Object `tfsdk:"search_criteria"`
}

func (r ArchivePolicyResourceModel) toAPIModel(ctx context.Context, apiModel *ArchivePolicyAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	attrs := r.SearchCriteria.Attributes()
	searchCriteria := ArchivePolicySearchCriteriaAPIModel{
		IncludeAllProjects:           attrs["include_all_projects"].(types.Bool).ValueBoolPointer(),
		CreatedBeforeInMonths:        attrs["created_before_in_months"].(types.Int64).ValueInt64Pointer(),
		LastDownloadedBeforeInMonths: attrs["last_downloaded_before_in_months"].(types.Int64).ValueInt64Pointer(),
		KeepLastNVerions:             attrs["keep_last_n_versions"].(types.Int64).ValueInt64Pointer(),
	}

	diags.Append(attrs["package_types"].(types.Set).ElementsAs(ctx, &searchCriteria.PackageTypes, false)...)
	diags.Append(attrs["repos"].(types.Set).ElementsAs(ctx, &searchCriteria.Repos, false)...)
	diags.Append(attrs["excluded_repos"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludedRepos, false)...)
	diags.Append(attrs["included_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedPackages, false)...)
	diags.Append(attrs["excluded_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludedPackages, false)...)
	diags.Append(attrs["included_projects"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedProjects, false)...)

	*apiModel = ArchivePolicyAPIModel{
		Key:               r.Key.ValueString(),
		Description:       r.Description.ValueString(),
		CronExpression:    r.CronExpression.ValueString(),
		DurationInMinutes: r.DurationInMinutes.ValueInt64(),
		SkipTrashcan:      r.SkipTrashcan.ValueBool(),
		ProjectKey:        r.ProjectKey.ValueString(),
		SearchCriteria:    searchCriteria,
	}

	return diags
}

func (r *ArchivePolicyResourceModel) fromAPIModel(ctx context.Context, apiModel ArchivePolicyAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.Key = types.StringValue(apiModel.Key)
	r.Description = types.StringValue(apiModel.Description)
	r.CronExpression = types.StringValue(apiModel.CronExpression)
	r.DurationInMinutes = types.Int64Value(apiModel.DurationInMinutes)
	r.Enabled = types.BoolValue(apiModel.Enabled)
	r.SkipTrashcan = types.BoolValue(apiModel.SkipTrashcan)

	packageTypes, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.PackageTypes)
	if ds.HasError() {
		diags.Append(ds...)
	}

	repos, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.Repos)
	if ds.HasError() {
		diags.Append(ds...)
	}

	excludedRepos := types.SetNull(types.StringType)
	if apiModel.SearchCriteria.ExcludedRepos != nil {
		set, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.ExcludedRepos)
		if ds.HasError() {
			diags.Append(ds...)
		}
		excludedRepos = set
	}

	includedPackages, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.IncludedPackages)
	if ds.HasError() {
		diags.Append(ds...)
	}

	excludedPackages := types.SetNull(types.StringType)
	if apiModel.SearchCriteria.ExcludedPackages != nil {
		set, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.ExcludedPackages)
		if ds.HasError() {
			diags.Append(ds...)
		}
		excludedPackages = set
	}

	includedProjects, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.IncludedProjects)
	if ds.HasError() {
		diags.Append(ds...)
	}

	searchCriteria, ds := types.ObjectValue(
		map[string]attr.Type{
			"package_types":                    types.SetType{ElemType: types.StringType},
			"repos":                            types.SetType{ElemType: types.StringType},
			"excluded_repos":                   types.SetType{ElemType: types.StringType},
			"included_packages":                types.SetType{ElemType: types.StringType},
			"excluded_packages":                types.SetType{ElemType: types.StringType},
			"include_all_projects":             types.BoolType,
			"included_projects":                types.SetType{ElemType: types.StringType},
			"created_before_in_months":         types.Int64Type,
			"last_downloaded_before_in_months": types.Int64Type,
			"keep_last_n_versions":             types.Int64Type,
		},
		map[string]attr.Value{
			"package_types":                    packageTypes,
			"repos":                            repos,
			"excluded_repos":                   excludedRepos,
			"included_packages":                includedPackages,
			"excluded_packages":                excludedPackages,
			"include_all_projects":             types.BoolPointerValue(apiModel.SearchCriteria.IncludeAllProjects),
			"included_projects":                includedProjects,
			"created_before_in_months":         types.Int64PointerValue(apiModel.SearchCriteria.CreatedBeforeInMonths),
			"last_downloaded_before_in_months": types.Int64PointerValue(apiModel.SearchCriteria.LastDownloadedBeforeInMonths),
			"keep_last_n_versions":             types.Int64PointerValue(apiModel.SearchCriteria.KeepLastNVerions),
		},
	)
	if ds.HasError() {
		diags.Append(ds...)
	}

	r.SearchCriteria = searchCriteria

	return diags
}

type ArchivePolicyAPIModel struct {
	Key               string                              `json:"key"`
	Description       string                              `json:"description,omitempty"`
	CronExpression    string                              `json:"cronExp"`
	DurationInMinutes int64                               `json:"durationInMinutes"`
	Enabled           bool                                `json:"enabled,omitempty"`
	SkipTrashcan      bool                                `json:"skipTrashcan"`
	ProjectKey        string                              `json:"projectKey"`
	SearchCriteria    ArchivePolicySearchCriteriaAPIModel `json:"searchCriteria"`
}

type ArchivePolicySearchCriteriaAPIModel struct {
	PackageTypes                 []string  `json:"packageTypes"`
	Repos                        []string  `json:"repos"`
	ExcludedRepos                *[]string `json:"excludedRepos,omitempty"`
	IncludedPackages             []string  `json:"includedPackages"`
	ExcludedPackages             *[]string `json:"excludedPackages,omitempty"`
	IncludeAllProjects           *bool     `json:"includeAllProjects,omitempty"`
	IncludedProjects             *[]string `json:"includedProjects,omitempty"`
	CreatedBeforeInMonths        *int64    `json:"createdBeforeInMonths,omitempty"`
	LastDownloadedBeforeInMonths *int64    `json:"lastDownloadedBeforeInMonths,omitempty"`
	KeepLastNVerions             *int64    `json:"keepLastNVerions,omitempty"`
}

type ArchivePolicyEnablementAPIModel struct {
	Enabled bool `json:"enabled"`
}

func (r *ArchivePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`), "only letters, numbers, underscore and hyphen are allowed"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "An ID that is used to identify the archive policy. A minimum of three characters is required and can include letters, numbers, underscore and hyphen.",
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"cron_expression": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorfw_string.IsCron(),
				},
				MarkdownDescription: "The cron expression determines when the policy is run. This parameter is not mandatory, however if left empty the policy will not run automatically and can only be triggered manually.",
			},
			"duration_in_minutes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum duration (in minutes) for policy execution, after which the policy will stop running even if not completed. While setting a maximum run duration for a policy is useful for adhering to a strict archive V2 schedule, it can cause the policy to stop before completion.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enables or disabled the package cleanup policy. This allows the user to run the policy manually. If a policy has a valid cron expression, then it will be scheduled for execution based on it. If a policy is disabled, its future executions will be unscheduled. Defaults to `true`",
			},
			"skip_trashcan": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "A `true` value means that when this policy is executed, packages will be permanently deleted. `false` means that when the policy is executed packages will be deleted to the Trash Can. Defaults to `false`.\n\n" +
					"~>The Global Trash Can setting must be enabled if you want deleted items to be transferred to the Trash Can. For information on enabling global Trash Can settings, see [Trash Can Settings](https://jfrog.com/help/r/jfrog-artifactory-documentation/trash-can-settings).",
			},
			"project_key": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorfw_string.ProjectKey(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "This attribute is used only for project-level archive V2 policies, it is not used for global-level policies.",
			},
			"search_criteria": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"package_types": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(archivePolicySupportedPackageType...),
							),
						},
						MarkdownDescription: fmt.Sprintf("The package types that are archived by the policy. Support: %s.", strings.Join(archivePolicySupportedPackageType, ", ")),
					},
					"repos": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						MarkdownDescription: "Specify one or more patterns for the repository name(s) on which you want the archive policy to run. You can also specify explicit repository names. Specifying at least one pattern or explicit name is required. Only packages in repositories that match the pattern or explicit name will be archived. For including all repos use `**`. Example: `repos = [\"**\"]`",
					},
					"excluded_repos": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						MarkdownDescription: "Specify patterns for repository names or explicit repository names that you want excluded from the archive policy.",
					},
					"included_packages": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Validators: []validator.Set{
							setvalidator.SizeBetween(1, 1),
						},
						MarkdownDescription: "Specify a pattern for a package name or an explicit package name. It accept only single element which can be specific package or pattern, and for including all packages use `**`. Example: `included_packages = [\"**\"]`",
					},
					"excluded_packages": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						MarkdownDescription: "Specify explicit package names that you want excluded from the policy. Only Name explicit names (and not patterns) are accepted.",
					},
					"include_all_projects": schema.BoolAttribute{
						Optional:    true,
						Description: "Set this value to `true` if you want the policy to run on all Artifactory projects. The default value is `false`.\n\n~>This attribute is relevant only on the global level, for Platform Admins.",
					},
					"included_projects": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						MarkdownDescription: "List of projects on which you want this policy to run. To include repositories that are not assigned to any project, enter the project key `default`.\n\n" +
							"~>This setting is relevant only on the global level, for Platform Admins.",
					},
					"created_before_in_months": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(24),
						Validators: []validator.Int64{
							int64validator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("last_downloaded_before_in_months")),
							int64validator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("keep_last_n_versions"),
							),
							int64validator.AtLeast(0),
						},
						MarkdownDescription: "The archive policy will archive packages based on how long ago they were created. For example, if this parameter is 2 then packages created more than 2 months ago will be archived as part of the policy.",
					},
					"last_downloaded_before_in_months": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(24),
						Validators: []validator.Int64{
							int64validator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("created_before_in_months")),
							int64validator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("keep_last_n_versions"),
							),
							int64validator.AtLeast(0),
						},
						MarkdownDescription: "The archive policy will archive packages based on how long ago they were downloaded. For example, if this parameter is 5 then packages downloaded more than 5 months ago will be archived as part of the policy.\n\n" +
							"~>JFrog recommends using the `last_downloaded_before_in_months` condition to ensure that packages currently in use are not archived.",
					},
					"keep_last_n_versions": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("created_before_in_months"),
								path.MatchRelative().AtParent().AtName("last_downloaded_before_in_months"),
							),
							int64validator.AtLeast(1),
						},
						MarkdownDescription: "Set a value for the number of latest versions to keep. The archive policy will remove all versions before the number you select here. The latest version is always excluded.\n\n" +
							"~>Versions are determined by creation date.\n\n" +
							"~>Not all package types support this condition. If you include a package type in your policy that is not compatible with this condition, a validation error (400) is returned. For information on which package types support this condition, see [here]().",
					},
				},
				Required: true,
			},
		},
		Description: "Provides an Artifactory Archive Policy resource. This resource enable system administrators to define and customize policies based on specific criteria for removing unused binaries from across their JFrog platform. " +
			"See [Archive Policies](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies) for more details.\n\n" +
			"~>Currently in beta and not yet globally available. A full rollout is scheduled for Q1 2025.",
	}
}

func (r ArchivePolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ArchivePolicyResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If search_criteria is not configured, return without warning.
	if data.SearchCriteria.IsNull() || data.SearchCriteria.IsUnknown() {
		return
	}

	searchCriteriaAttrs := data.SearchCriteria.Attributes()
	createdBeforeInMonths := searchCriteriaAttrs["created_before_in_months"].(types.Int64)
	lastDownloadedBeforeInMonths := searchCriteriaAttrs["last_downloaded_before_in_months"].(types.Int64)

	if createdBeforeInMonths.IsNull() || createdBeforeInMonths.IsUnknown() {
		return
	}

	if lastDownloadedBeforeInMonths.IsNull() || lastDownloadedBeforeInMonths.IsUnknown() {
		return
	}

	if createdBeforeInMonths.ValueInt64() == 0 && lastDownloadedBeforeInMonths.ValueInt64() == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("search_criteria").AtName("created_before_in_months"),
			"Invalid Attribute Configuration",
			"Both created_before_in_months and last_downloaded_before_in_months cannot be zero at the same time.",
		)
	}
}

func (r *ArchivePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArchivePolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy ArchivePolicyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		SetError(&jfrogErrors).
		Post(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	if plan.Enabled.ValueBool() {
		policyEnablement := ArchivePolicyEnablementAPIModel{
			Enabled: true,
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			SetError(&jfrogErrors).
			Post(r.EnablementEndpoint)

		if enablementErr != nil {
			utilfw.UnableToCreateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArchivePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArchivePolicyResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var policy ArchivePolicyAPIModel
	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		SetResult(&policy).
		SetError(&jfrogErrors).
		Get(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, jfrogErrors.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ArchivePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArchivePolicyResourceModel
	var state ArchivePolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy ArchivePolicyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// policy.Enabled can't be changed using update API so set the field to
	// the current state's value
	policy.Enabled = state.Enabled.ValueBool()

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		SetError(&jfrogErrors).
		Put(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	enabledChanged := state.Enabled.ValueBool() != plan.Enabled.ValueBool()
	if enabledChanged {
		policyEnablement := ArchivePolicyEnablementAPIModel{}
		if state.Enabled.ValueBool() && !plan.Enabled.ValueBool() { // if Enabled goes from true to false
			policyEnablement.Enabled = false
		} else if !state.Enabled.ValueBool() && plan.Enabled.ValueBool() { // if Enabled goes from false to true
			policyEnablement.Enabled = true
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			SetError(&jfrogErrors).
			Post(r.EnablementEndpoint)

		if enablementErr != nil {
			utilfw.UnableToUpdateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArchivePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArchivePolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		SetError(&jfrogErrors).
		Delete(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, jfrogErrors.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArchivePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	if len(parts) > 0 && parts[0] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[0])...)
	}

	if len(parts) == 2 && parts[1] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[1])...)
	}
}
