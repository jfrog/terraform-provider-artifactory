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
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	PackageCleanupPolicyEndpointPath           = "artifactory/api/cleanup/packages/policies/{policyKey}"
	PackageCleanupPolicyEnablementEndpointPath = "artifactory/api/cleanup/packages/policies/{policyKey}/enablement"
)

var cleanupPolicySupportedPackageType = []string{"conan", "docker", "generic", "gradle", "maven", "npm", "nuget", "rpm"}

func NewPackageCleanupPolicyResource() resource.Resource {
	return &PackageCleanupPolicyResource{
		TypeName: "artifactory_package_cleanup_policy",
	}
}

type PackageCleanupPolicyResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type PackageCleanupPolicyResourceModelV0 struct {
	Key               types.String `tfsdk:"key"`
	Description       types.String `tfsdk:"description"`
	CronExpression    types.String `tfsdk:"cron_expression"`
	DurationInMinutes types.Int64  `tfsdk:"duration_in_minutes"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SkipTrashcan      types.Bool   `tfsdk:"skip_trashcan"`
	SearchCriteria    types.Object `tfsdk:"search_criteria"`
}

type PackageCleanupPolicyResourceModelV1 struct {
	PackageCleanupPolicyResourceModelV0
	ProjectKey types.String `tfsdk:"project_key"`
}

func (r PackageCleanupPolicyResourceModelV1) toAPIModel(ctx context.Context, apiModel *PackageCleanupPolicyAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	attrs := r.SearchCriteria.Attributes()
	searchCriteria := PackageCleanupPolicySearchCriteriaAPIModel{
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

	*apiModel = PackageCleanupPolicyAPIModel{
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

func (r *PackageCleanupPolicyResourceModelV1) fromAPIModel(ctx context.Context, apiModel PackageCleanupPolicyAPIModel) diag.Diagnostics {
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

	includeAllProjects := types.BoolNull()
	if apiModel.SearchCriteria.IncludeAllProjects != nil {
		includeAllProjects = types.BoolPointerValue(apiModel.SearchCriteria.IncludeAllProjects)
	}

	createdBeforeInMonths := types.Int64Null()
	if apiModel.SearchCriteria.CreatedBeforeInMonths != nil {
		createdBeforeInMonths = types.Int64PointerValue(apiModel.SearchCriteria.CreatedBeforeInMonths)
	}

	lastDownloadedBeforeInMonths := types.Int64Null()
	if apiModel.SearchCriteria.LastDownloadedBeforeInMonths != nil {
		lastDownloadedBeforeInMonths = types.Int64PointerValue(apiModel.SearchCriteria.LastDownloadedBeforeInMonths)
	}

	keepLastNVerions := types.Int64Null()
	if apiModel.SearchCriteria.KeepLastNVerions != nil {
		keepLastNVerions = types.Int64PointerValue(apiModel.SearchCriteria.KeepLastNVerions)
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
			"include_all_projects":             includeAllProjects,
			"included_projects":                includedProjects,
			"created_before_in_months":         createdBeforeInMonths,
			"last_downloaded_before_in_months": lastDownloadedBeforeInMonths,
			"keep_last_n_versions":             keepLastNVerions,
		},
	)
	if ds.HasError() {
		diags.Append(ds...)
	}

	r.SearchCriteria = searchCriteria

	return diags
}

type PackageCleanupPolicyAPIModel struct {
	Key               string                                     `json:"key"`
	Description       string                                     `json:"description,omitempty"`
	CronExpression    string                                     `json:"cronExp"`
	DurationInMinutes int64                                      `json:"durationInMinutes"`
	Enabled           bool                                       `json:"enabled,omitempty"`
	SkipTrashcan      bool                                       `json:"skipTrashcan"`
	ProjectKey        string                                     `json:"projectKey"`
	SearchCriteria    PackageCleanupPolicySearchCriteriaAPIModel `json:"searchCriteria"`
}

type PackageCleanupPolicySearchCriteriaAPIModel struct {
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

type PackageCleanupPolicyEnablementAPIModel struct {
	Enabled bool `json:"enabled"`
}

func (r *PackageCleanupPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

var cleanupPolicySchemaV0 = map[string]schema.Attribute{
	"key": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: "Policy key. It has to be unique. It should not be used for other policies and configuration entities like archive policies, key pairs, repo layouts, property sets, backups, proxies, reverse proxies etc.",
	},
	"description": schema.StringAttribute{
		Optional: true,
	},
	"cron_expression": schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			validatorfw_string.IsCron(),
		},
		MarkdownDescription: "The Cron expression that sets the schedule of policy execution. For example, `0 0 2 * * ?` executes the policy every day at 02:00 AM. The minimum recurrent time for policy execution is 6 hours.",
	},
	"duration_in_minutes": schema.Int64Attribute{
		Optional:            true,
		MarkdownDescription: "Enable and select the maximum duration for policy execution. Note: using this setting can cause the policy to stop before completion.",
	},
	"enabled": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
		MarkdownDescription: "Enables or disabled the package cleanup policy. This allows the user to run the policy manually. If a policy has a valid cron expression, then it will be scheduled for execution based on it. If a policy is disabled, its future executions will be unscheduled. Defaults to `true`",
	},
	"skip_trashcan": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: "When enabled, deleted packages are permanently removed from Artifactory without an option to restore them. Defaults to `false`",
	},
	"search_criteria": schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"package_types": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(cleanupPolicySupportedPackageType...),
					),
				},
				MarkdownDescription: fmt.Sprintf("Types of packages to be removed. Support: %s.", strings.Join(cleanupPolicySupportedPackageType, ", ")),
			},
			"repos": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "Specify patterns for repository names or explicit repository names. For including all repos use `**`. Example: `repos = [\"**\"]`",
			},
			"excluded_repos": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "Specify patterns for repository names or explicit repository names that you want excluded from the policy. It can not accept any pattern only list of specific repositories.",
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
				MarkdownDescription: "Specify explicit package names that you want excluded from the policy.",
			},
			"include_all_projects": schema.BoolAttribute{
				Optional: true,
			},
			"included_projects": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "List of projects name(s) to apply the policy to.",
			},
			"created_before_in_months": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("last_downloaded_before_in_months")),
					int64validator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("keep_last_n_versions"),
					),
					int64validator.AtLeast(1),
				},
				MarkdownDescription: "Remove packages based on when they were created.",
			},
			"last_downloaded_before_in_months": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("created_before_in_months")),
					int64validator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("keep_last_n_versions"),
					),
					int64validator.AtLeast(1),
				},
				MarkdownDescription: "Remove packages based on when they were last downloaded.",
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
				MarkdownDescription: "Select the number of latest version to keep. The policy will remove all versions (based on creation date) prior to the selected number. Some package types may not be supported. [Learn more](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies/package-types-coverage)",
			},
		},
		Required: true,
	},
}

var cleanupPolicySchemaV1 = lo.Assign(
	cleanupPolicySchemaV0,
	map[string]schema.Attribute{
		"key": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(3),
				stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`), "only letters, numbers, underscore and hyphen are allowed"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: "Policy key. It has to be unique. It should not be used for other policies and configuration entities like archive policies, key pairs, repo layouts, property sets, backups, proxies, reverse proxies etc. A minimum of three characters is required and can include letters, numbers, underscore and hyphen.",
		},
		"project_key": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				validatorfw_string.ProjectKey(),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "This attribute is used only for project-level cleanup policies, it is not used for global-level policies.",
		},
		"skip_trashcan": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			MarkdownDescription: "Enabling this setting results in packages being permanently deleted from Artifactory after the cleanup policy is executed instead of going to the Trash Can repository. Defaults to `false`.\n\n" +
				"~>The Global Trash Can setting must be enabled if you want deleted items to be transferred to the Trash Can. For information on enabling global Trash Can settings, see [Trash Can Settings](https://jfrog.com/help/r/jfrog-artifactory-documentation/trash-can-settings).",
		},
		"search_criteria": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"package_types": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.Set{
						setvalidator.ValueStringsAre(
							stringvalidator.OneOf(cleanupPolicySupportedPackageType...),
						),
					},
					MarkdownDescription: fmt.Sprintf("Types of packages to be removed. Support: %s.", strings.Join(cleanupPolicySupportedPackageType, ", ")),
				},
				"repos": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					MarkdownDescription: "Specify patterns for repository names or explicit repository names. For including all repos use `**`. Example: `repos = [\"**\"]`",
				},
				"excluded_repos": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					MarkdownDescription: "Specify patterns for repository names or explicit repository names that you want excluded from the cleanup policy.",
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
					Description: "Set this to `true` if you want the policy to run on all projects on the platform.",
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
						int64validator.AtLeast(1),
					},
					MarkdownDescription: "Remove packages based on when they were created. For example, remove packages that were created more than a year ago. The default value is to remove packages created more than 2 years ago.",
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
						int64validator.AtLeast(1),
					},
					MarkdownDescription: "Removes packages based on when they were last downloaded. For example, removes packages that were not downloaded in the past year. The default value is to remove packages that were downloaded more than 2 years ago.\n\n" +
						"~>If a package was never downloaded, the policy will remove it based only on the age-condition (`created_before_in_months`).\n\n" +
						"~>JFrog recommends using the `last_downloaded_before_in_months` condition to ensure that packages currently in use are not deleted.",
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
					MarkdownDescription: "Select the number of latest versions to keep. The cleanup policy will remove all versions prior to the number you select here. The latest version is always excluded. Versions are determined by creation date.\n\n" +
						"~>Not all package types support this condition. For information on which package types support this condition, [learn more](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies/package-types-coverage).",
				},
			},
			Required: true,
		},
	},
)

func (r *PackageCleanupPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: cleanupPolicySchemaV1,
		Version:    1,
		Description: "Provides an Artifactory Package Cleanup Policy resource. This resource enable system administrators to define and customize policies based on specific criteria for removing unused binaries from across their JFrog platform. " +
			"See [Rentation Policies](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies) for more details.\n\n" +
			"~>Currently in beta and will be globally available in v7.98.x.",
	}
}

func (r *PackageCleanupPolicyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: cleanupPolicySchemaV0,
			},
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData PackageCleanupPolicyResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := PackageCleanupPolicyResourceModelV1{
					PackageCleanupPolicyResourceModelV0: priorStateData,
					ProjectKey:                          types.StringNull(),
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

func (r *PackageCleanupPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *PackageCleanupPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan PackageCleanupPolicyResourceModelV1

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy PackageCleanupPolicyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		Post(PackageCleanupPolicyEndpointPath)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	if plan.Enabled.ValueBool() {
		policyEnablement := PackageCleanupPolicyEnablementAPIModel{
			Enabled: true,
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			Post(PackageCleanupPolicyEnablementEndpointPath)

		if enablementErr != nil {
			utilfw.UnableToCreateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToCreateResourceError(resp, enablementResp.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageCleanupPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state PackageCleanupPolicyResourceModelV1
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var policy PackageCleanupPolicyAPIModel

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		SetResult(&policy).
		Get(PackageCleanupPolicyEndpointPath)

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
		utilfw.UnableToRefreshResourceError(resp, response.String())
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

func (r *PackageCleanupPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan PackageCleanupPolicyResourceModelV1
	var state PackageCleanupPolicyResourceModelV1

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

	var policy PackageCleanupPolicyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// policy.Enabled can't be changed using update API so set the field to
	// the current state's value
	policy.Enabled = state.Enabled.ValueBool()

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		Put(PackageCleanupPolicyEndpointPath)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	enabledChanged := state.Enabled.ValueBool() != plan.Enabled.ValueBool()
	if enabledChanged {
		policyEnablement := PackageCleanupPolicyEnablementAPIModel{}
		if state.Enabled.ValueBool() && !plan.Enabled.ValueBool() { // if Enabled goes from true to false
			policyEnablement.Enabled = false
		} else if !state.Enabled.ValueBool() && plan.Enabled.ValueBool() { // if Enabled goes from false to true
			policyEnablement.Enabled = true
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			Post(PackageCleanupPolicyEnablementEndpointPath)

		if enablementErr != nil {
			utilfw.UnableToUpdateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToUpdateResourceError(resp, enablementResp.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageCleanupPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state PackageCleanupPolicyResourceModelV1

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		Delete(PackageCleanupPolicyEndpointPath)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *PackageCleanupPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	if len(parts) > 0 && parts[0] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[0])...)
	}

	if len(parts) == 2 && parts[1] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[1])...)
	}
}
