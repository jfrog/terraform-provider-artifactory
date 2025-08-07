package configuration

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

// Custom validator for search_criteria to enforce validation rules
type cleanupSearchCriteriaValidator struct{}

func (v cleanupSearchCriteriaValidator) Description(ctx context.Context) string {
	return "Validates that exactly one group of conditions is specified (time-based or version-based)"
}

func (v cleanupSearchCriteriaValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that exactly one group of conditions is specified (time-based or version-based)"
}

func (v cleanupSearchCriteriaValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Get the object value
	obj := req.ConfigValue

	// If the object is null or unknown, skip validation
	if obj.IsNull() || obj.IsUnknown() {
		return
	}

	// Get the attributes
	attrs := obj.Attributes()

	// Helper function to get int64 value
	getInt64 := func(key string) types.Int64 {
		if v, ok := attrs[key]; ok && !v.IsNull() && !v.IsUnknown() {
			if val, ok := v.(types.Int64); ok {
				return val
			}
		}
		return types.Int64Null()
	}

	// Check for time-based conditions (days) - for Artifactory 7.111.2+
	createdBeforeInDays := getInt64("created_before_in_days")
	lastDownloadedBeforeInDays := getInt64("last_downloaded_before_in_days")

	// Check for time-based conditions (months) - for Artifactory < 7.111.2
	createdBeforeInMonths := getInt64("created_before_in_months")
	lastDownloadedBeforeInMonths := getInt64("last_downloaded_before_in_months")

	// Version-based condition (available in both versions)
	keepLastNVersions := getInt64("keep_last_n_versions")

	// Helper function to check if properties are set
	checkPropertiesSet := func(key string) bool {
		if v, ok := attrs[key]; ok && !v.IsNull() && !v.IsUnknown() {
			if m, ok := v.(types.Map); ok {
				return len(m.Elements()) > 0
			}
		}
		return false
	}

	// Check if days attributes are set (7.111.2+)
	createdDaysSet := !createdBeforeInDays.IsNull() && !createdBeforeInDays.IsUnknown() && createdBeforeInDays.ValueInt64() > 0
	downloadedDaysSet := !lastDownloadedBeforeInDays.IsNull() && !lastDownloadedBeforeInDays.IsUnknown() && lastDownloadedBeforeInDays.ValueInt64() > 0
	timeBasedDaysSet := createdDaysSet || downloadedDaysSet

	// Check if months attributes are set (< 7.111.2)
	createdMonthsSet := !createdBeforeInMonths.IsNull() && !createdBeforeInMonths.IsUnknown() && createdBeforeInMonths.ValueInt64() > 0
	downloadedMonthsSet := !lastDownloadedBeforeInMonths.IsNull() && !lastDownloadedBeforeInMonths.IsUnknown() && lastDownloadedBeforeInMonths.ValueInt64() > 0
	timeBasedMonthsSet := createdMonthsSet || downloadedMonthsSet

	// Version-based condition
	keepVersionBasedSet := !keepLastNVersions.IsNull() && !keepLastNVersions.IsUnknown() && keepLastNVersions.ValueInt64() > 0

	// Check for zero values in time-based conditions
	if (!createdBeforeInDays.IsNull() && !createdBeforeInDays.IsUnknown() && createdBeforeInDays.ValueInt64() == 0) ||
		(!lastDownloadedBeforeInDays.IsNull() && !lastDownloadedBeforeInDays.IsUnknown() && lastDownloadedBeforeInDays.ValueInt64() == 0) ||
		(!createdBeforeInMonths.IsNull() && !createdBeforeInMonths.IsUnknown() && createdBeforeInMonths.ValueInt64() == 0) ||
		(!lastDownloadedBeforeInMonths.IsNull() && !lastDownloadedBeforeInMonths.IsUnknown() && lastDownloadedBeforeInMonths.ValueInt64() == 0) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Configuration",
			"Time-based conditions must have a value greater than 0. Zero values are not allowed for `created_before_in_days`, `last_downloaded_before_in_days`, `created_before_in_months`, or `last_downloaded_before_in_months`.",
		)
		return
	}

	// Check for zero values in version-based condition
	if !keepLastNVersions.IsNull() && !keepLastNVersions.IsUnknown() && keepLastNVersions.ValueInt64() == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Configuration",
			"Version-based condition (keep_last_n_versions) must have a value greater than 0. Zero values are not allowed.",
		)
		return
	}

	// Properties-based conditions (only included_properties)
	includedPropertiesSet := checkPropertiesSet("included_properties")
	propertiesBasedSet := includedPropertiesSet

	// Check for mixed usage of days and months (invalid)
	if timeBasedDaysSet && timeBasedMonthsSet {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Configuration",
			"Cannot use both days-based conditions (`created_before_in_days`, `last_downloaded_before_in_days`) and months-based conditions (`created_before_in_months`, `last_downloaded_before_in_months`) together. Use either days-based or months-based conditions based on your Artifactory version.",
		)
		return
	}

	// Check for time-based conditions (either days or months)
	timeBasedSet := timeBasedDaysSet || timeBasedMonthsSet

	// Count how many different condition types are set
	conditionTypes := 0
	if timeBasedSet {
		conditionTypes++
	}
	if keepVersionBasedSet {
		conditionTypes++
	}
	if propertiesBasedSet {
		conditionTypes++
	}

	// Must specify at least one condition
	if conditionTypes == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Configuration",
			"A policy must use exactly one of the following condition types: time-based conditions (days-based or months-based), version-based condition (keep_last_n_versions), or properties-based condition (included_properties). Cannot use multiple condition types together.",
		)
		return
	}

	// Cannot use multiple condition types together
	if conditionTypes > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Policy Configuration",
			"A policy can only use one type of condition: either time-based conditions (days-based or months-based), version-based condition (keep_last_n_versions), or properties-based condition (included_properties). Cannot use multiple condition types together.",
		)
		return
	}
}

type cleanupSingleKeySingleValueMapValidator struct{}

func (v cleanupSingleKeySingleValueMapValidator) Description(ctx context.Context) string {
	return "Must have exactly one key and that key must have exactly one string value"
}

func (v cleanupSingleKeySingleValueMapValidator) MarkdownDescription(ctx context.Context) string {
	return "Must have exactly one key and that key must have exactly one string value"
}

func (v cleanupSingleKeySingleValueMapValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	m := req.ConfigValue.Elements()
	if len(m) != 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Properties",
			"Properties-based conditions must have exactly one key.",
		)
		return
	}

	for _, v := range m {
		if v.IsNull() || v.IsUnknown() {
			continue
		}
		if l, ok := v.(types.List); ok {
			if len(l.Elements()) != 1 {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid Properties",
					"The property value must be a list with exactly one string value.",
				)
			}
		}
	}
}

const (
	PackageCleanupPolicyEndpointPath           = "artifactory/api/cleanup/packages/policies/{policyKey}"
	PackageCleanupPolicyEnablementEndpointPath = "artifactory/api/cleanup/packages/policies/{policyKey}/enablement"
)

var cleanupPolicySupportedPackageType = []string{
	repository.AlpinePackageType,
	repository.AnsiblePackageType,
	repository.CargoPackageType,
	repository.ChefPackageType,
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.ConanPackageType,
	repository.CondaPackageType,
	repository.DebianPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HelmPackageType,
	repository.HelmOCIPackageType,
	repository.HuggingFacePackageType,
	repository.MachineLearningType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.OCIPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.SBTPackageType,
	repository.SwiftPackageType,
	repository.TerraformPackageType,
	repository.TerraformBackendPackageType,
	repository.RPMPackageType,
	yumPolicyPackageType,
}

func NewPackageCleanupPolicyResource() resource.Resource {
	return &PackageCleanupPolicyResource{
		JFrogResource: util.JFrogResource{
			TypeName:         "artifactory_package_cleanup_policy",
			DocumentEndpoint: "artifactory/api/cleanup/packages/policies/{policyKey}",
		},
		EnablementEndpoint: "artifactory/api/cleanup/packages/policies/{policyKey}/enablement",
	}
}

var _ resource.Resource = (*PackageCleanupPolicyResource)(nil)

type PackageCleanupPolicyResource struct {
	util.JFrogResource
	EnablementEndpoint string
	ProviderData       util.ProviderMetadata
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

	// Helper function to safely get int64 pointer
	getInt64Pointer := func(key string) *int64 {
		if v, ok := attrs[key]; ok && !v.IsNull() && !v.IsUnknown() {
			if val, ok := v.(types.Int64); ok {
				return val.ValueInt64Pointer()
			}
		}
		return nil
	}

	searchCriteria := PackageCleanupPolicySearchCriteriaAPIModel{
		IncludeAllProjects:           attrs["include_all_projects"].(types.Bool).ValueBoolPointer(),
		CreatedBeforeInMonths:        getInt64Pointer("created_before_in_months"),
		LastDownloadedBeforeInMonths: getInt64Pointer("last_downloaded_before_in_months"),
		CreatedBeforeInDays:          getInt64Pointer("created_before_in_days"),
		LastDownloadedBeforeInDays:   getInt64Pointer("last_downloaded_before_in_days"),
		KeepLastNVersions:            getInt64Pointer("keep_last_n_versions"),
	}

	diags.Append(attrs["package_types"].(types.Set).ElementsAs(ctx, &searchCriteria.PackageTypes, false)...)
	diags.Append(attrs["repos"].(types.Set).ElementsAs(ctx, &searchCriteria.Repos, false)...)
	diags.Append(attrs["excluded_repos"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludedRepos, false)...)
	diags.Append(attrs["included_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedPackages, false)...)
	diags.Append(attrs["excluded_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludedPackages, false)...)
	diags.Append(attrs["included_projects"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedProjects, false)...)

	if v, ok := attrs["included_properties"]; ok && !v.IsNull() && !v.IsUnknown() {
		if m, ok := v.(types.Map); ok {
			searchCriteria.IncludedProperties = make(map[string][]string)
			for k, val := range m.Elements() {
				if l, ok := val.(types.List); ok && !l.IsNull() && !l.IsUnknown() {
					var values []string
					for _, lv := range l.Elements() {
						if s, ok := lv.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
							values = append(values, s.ValueString())
						}
					}
					searchCriteria.IncludedProperties[k] = values
				}
			}
		}
	}
	if v, ok := attrs["excluded_properties"]; ok && !v.IsNull() && !v.IsUnknown() {
		if m, ok := v.(types.Map); ok {
			searchCriteria.ExcludedProperties = make(map[string][]string)
			for k, val := range m.Elements() {
				if l, ok := val.(types.List); ok && !l.IsNull() && !l.IsUnknown() {
					var values []string
					for _, lv := range l.Elements() {
						if s, ok := lv.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
							values = append(values, s.ValueString())
						}
					}
					searchCriteria.ExcludedProperties[k] = values
				}
			}
		}
	}

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

	includedProperties := types.MapNull(types.ListType{ElemType: types.StringType})
	if apiModel.SearchCriteria.IncludedProperties != nil {
		m := map[string]attr.Value{}
		for k, v := range apiModel.SearchCriteria.IncludedProperties {
			lv, ds := types.ListValueFrom(ctx, types.StringType, v)
			if ds.HasError() {
				diags.Append(ds...)
			}
			m[k] = lv
		}
		includedProperties, _ = types.MapValue(types.ListType{ElemType: types.StringType}, m)
	}

	excludedProperties := types.MapNull(types.ListType{ElemType: types.StringType})
	if apiModel.SearchCriteria.ExcludedProperties != nil {
		m := map[string]attr.Value{}
		for k, v := range apiModel.SearchCriteria.ExcludedProperties {
			lv, ds := types.ListValueFrom(ctx, types.StringType, v)
			if ds.HasError() {
				diags.Append(ds...)
			}
			m[k] = lv
		}
		excludedProperties, _ = types.MapValue(types.ListType{ElemType: types.StringType}, m)
	}

	// Handle time-based attributes with proper null checking
	createdBeforeInMonths := types.Int64Null()
	if apiModel.SearchCriteria.CreatedBeforeInMonths != nil {
		createdBeforeInMonths = types.Int64PointerValue(apiModel.SearchCriteria.CreatedBeforeInMonths)
	}

	lastDownloadedBeforeInMonths := types.Int64Null()
	if apiModel.SearchCriteria.LastDownloadedBeforeInMonths != nil {
		lastDownloadedBeforeInMonths = types.Int64PointerValue(apiModel.SearchCriteria.LastDownloadedBeforeInMonths)
	}

	// Always set day-based attributes to ensure they are known values
	createdBeforeInDays := types.Int64Null()
	if apiModel.SearchCriteria.CreatedBeforeInDays != nil {
		createdBeforeInDays = types.Int64PointerValue(apiModel.SearchCriteria.CreatedBeforeInDays)
	}

	lastDownloadedBeforeInDays := types.Int64Null()
	if apiModel.SearchCriteria.LastDownloadedBeforeInDays != nil {
		lastDownloadedBeforeInDays = types.Int64PointerValue(apiModel.SearchCriteria.LastDownloadedBeforeInDays)
	}

	keepLastNVersions := types.Int64Null()
	if apiModel.SearchCriteria.KeepLastNVersions != nil {
		keepLastNVersions = types.Int64PointerValue(apiModel.SearchCriteria.KeepLastNVersions)
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
			"created_before_in_days":           types.Int64Type,
			"last_downloaded_before_in_days":   types.Int64Type,
			"keep_last_n_versions":             types.Int64Type,
			"excluded_properties":              types.MapType{ElemType: types.ListType{ElemType: types.StringType}},
			"included_properties":              types.MapType{ElemType: types.ListType{ElemType: types.StringType}},
		},
		map[string]attr.Value{
			"package_types":                    packageTypes,
			"repos":                            repos,
			"excluded_repos":                   excludedRepos,
			"included_packages":                includedPackages,
			"excluded_packages":                excludedPackages,
			"include_all_projects":             types.BoolPointerValue(apiModel.SearchCriteria.IncludeAllProjects),
			"included_projects":                includedProjects,
			"created_before_in_months":         createdBeforeInMonths,
			"last_downloaded_before_in_months": lastDownloadedBeforeInMonths,
			"created_before_in_days":           createdBeforeInDays,
			"last_downloaded_before_in_days":   lastDownloadedBeforeInDays,
			"keep_last_n_versions":             keepLastNVersions,
			"excluded_properties":              excludedProperties,
			"included_properties":              includedProperties,
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
	PackageTypes                 []string            `json:"packageTypes"`
	Repos                        []string            `json:"repos"`
	ExcludedRepos                *[]string           `json:"excludedRepos,omitempty"`
	IncludedPackages             []string            `json:"includedPackages"`
	ExcludedPackages             *[]string           `json:"excludedPackages,omitempty"`
	IncludeAllProjects           *bool               `json:"includeAllProjects"`
	IncludedProjects             *[]string           `json:"includedProjects,omitempty"`
	CreatedBeforeInMonths        *int64              `json:"createdBeforeInMonths,omitempty"`
	LastDownloadedBeforeInMonths *int64              `json:"lastDownloadedBeforeInMonths,omitempty"`
	CreatedBeforeInDays          *int64              `json:"createdBeforeInDays,omitempty"`
	LastDownloadedBeforeInDays   *int64              `json:"lastDownloadedBeforeInDays,omitempty"`
	KeepLastNVersions            *int64              `json:"keepLastNVersions,omitempty"`
	ExcludedProperties           map[string][]string `json:"excludedProperties,omitempty"`
	IncludedProperties           map[string][]string `json:"includedProperties,omitempty"`
}

type PackageCleanupPolicyEnablementAPIModel struct {
	Enabled bool `json:"enabled"`
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
		MarkdownDescription: "The cron expression that determines when the policy is run, However if left empty the policy will not run automatically and can only be triggered manually.",
	},
	"duration_in_minutes": schema.Int64Attribute{
		Optional:            true,
		MarkdownDescription: "The maximum duration (in minutes) for policy execution, after which the policy will stop running even if not completed. While setting a maximum run duration for a policy is useful for adhering to a strict cleanup schedule, it can cause the policy to stop before completion.",
	},
	"enabled": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
		MarkdownDescription: "A cleanup policy must be created inactive. But if used it must be set to `false`. If set to `true` when calling this API, the API call will fail and an error message is received.",
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
				Optional:            true,
				MarkdownDescription: "Remove packages based on when they were created.",
			},
			"last_downloaded_before_in_months": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Remove packages based on when they were last downloaded.",
			},
			"keep_last_n_versions": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Select the number of latest version to keep. The policy will remove all versions (based on creation date) prior to the selected number. Some package types may not be supported. [Learn more](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies/package-types-coverage)",
			},
			"created_before_in_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were created. For example, if this parameter is 5 then packages created more than 5 days ago will be deleted as part of the policy.\n\n" +
					"~>JFrog recommends using the `created_before_in_days` condition to ensure that packages currently in use are not deleted.",
			},
			"last_downloaded_before_in_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were downloaded. For example, if this parameter is 5 then packages downloaded more than 5 days ago will be deleted as part of the policy.\n\n" +
					"~>JFrog recommends using the `last_downloaded_before_in_days` condition to ensure that packages currently in use are not deleted.",
			},
			"excluded_properties": schema.MapAttribute{
				ElementType: types.ListType{ElemType: types.StringType},
				Optional:    true,
				Validators: []validator.Map{
					cleanupSingleKeySingleValueMapValidator{},
				},
				MarkdownDescription: "A key-value pair applied to the lead artifact of a package. Packages with this property will be excluded from deletion.",
			},
			"included_properties": schema.MapAttribute{
				ElementType: types.ListType{ElemType: types.StringType},
				Optional:    true,
				Validators: []validator.Map{
					cleanupSingleKeySingleValueMapValidator{},
				},
				MarkdownDescription: "A key-value pair applied to the lead artifact of a package. Packages with this property will be deleted.",
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
			Description: "An ID that is used to identify the cleanup policy. A minimum of three characters is required and can include letters, numbers, underscore and hyphen.",
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
			MarkdownDescription: "A true value means that when this policy is executed, packages will be permanently deleted. false means that when the policy is executed packages will be deleted to the Trash Can.\n\n" +
				"~>The Global Trash Can setting must be enabled if you want deleted items to be transferred to the Trash Can, see [Trash Can Settings](https://jfrog.com/help/r/jfrog-artifactory-documentation/trash-can-settings).",
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
					MarkdownDescription: fmt.Sprintf("The package types that are cleaned up by the policy. Support: %s.", strings.Join(cleanupPolicySupportedPackageType, ", ")),
				},
				"repos": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					MarkdownDescription: "Specify one or more patterns for the repository name(s) on which you want the cleanup policy to run. You can also specify explicit repository names. Specifying at least one pattern or explicit name is mandatory. Only packages in repositories that match the pattern or explicit name will be deleted. For including all repos use `**`. Example: `repos = [\"**\"]`",
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
					MarkdownDescription: "Specify a pattern for a package name or an explicit package name on which you want the cleanup policy to run. Only one pattern or explicit name can be entered. To include all packages, use `**`. Example: `included_packages = [\"**\"]`",
				},
				"excluded_packages": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					MarkdownDescription: "Specify explicit package names that you want excluded from the policy. Only explicit names (and not patterns) are accepted.",
				},
				"include_all_projects": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(false),
					MarkdownDescription: "Set this value to `true` if you want the policy to run on all Artifactory projects. The default value is `false`.\n\n " +
						"~>This parameter is relevant only on the global level, for Platform Admins.",
				},
				"included_projects": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(0),
					},
					MarkdownDescription: "Enter the project keys for the projects on which you want the policy to run. To include repositories that are not assigned to any project, enter the project key `default`.\n\n" +
						"~>This parameter is relevant only on the global level, for Platform Admins.",
				},
				"created_before_in_months": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were created. For example, if this parameter is 2 then packages created more than 2 months ago will be deleted as part of the policy.\n\n" +
						"~>JFrog recommends using the `created_before_in_months` condition to ensure that packages currently in use are not deleted.",
					DeprecationMessage: "Use `created_before_in_days` instead of `created_before_in_months`. Renamed to `created_before_in_days` starting in version 7.111.2.",
				},
				"last_downloaded_before_in_months": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were downloaded. For example, if this parameter is 5 then packages downloaded more than 5 months ago will be deleted as part of the policy.\n\n" +
						"~>JFrog recommends using the `last_downloaded_before_in_months` condition to ensure that packages currently in use are not deleted.",
					DeprecationMessage: "Use `last_downloaded_before_in_days` instead of `last_downloaded_before_in_months`. Renamed to `last_downloaded_before_in_days` starting in version 7.111.2.",
				},
				"created_before_in_days": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were created. For example, if this parameter is 5 then packages created more than 5 days ago will be deleted as part of the policy.\n\n" +
						"~>JFrog recommends using the `created_before_in_days` condition to ensure that packages currently in use are not deleted.",
				},
				"last_downloaded_before_in_days": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "The cleanup policy will delete packages based on how long ago they were downloaded. For example, if this parameter is 5 then packages downloaded more than 5 days ago will be deleted as part of the policy.\n\n" +
						"~>JFrog recommends using the `last_downloaded_before_in_days` condition to ensure that packages currently in use are not deleted.",
				},
				"keep_last_n_versions": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Set a value for the number of latest versions to keep. The cleanup policy will remove all versions prior to the number you select here. The latest version is always excluded.\n\n" +
						"~>Not all package types support this condition. For information on which package types support this condition, [learn more](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies/package-types-coverage).",
				},
				"excluded_properties": schema.MapAttribute{
					ElementType: types.ListType{ElemType: types.StringType},
					Optional:    true,
					Validators: []validator.Map{
						cleanupSingleKeySingleValueMapValidator{},
					},
					MarkdownDescription: "A key-value pair applied to the lead artifact of a package. Packages with this property will be excluded from deletion.",
				},
				"included_properties": schema.MapAttribute{
					ElementType: types.ListType{ElemType: types.StringType},
					Optional:    true,
					Validators: []validator.Map{
						cleanupSingleKeySingleValueMapValidator{},
					},
					MarkdownDescription: "A key-value pair applied to the lead artifact of a package. Packages with this property will be deleted.",
				},
			},
			Required: true,
			Validators: []validator.Object{
				cleanupSearchCriteriaValidator{},
			},
		},
	},
)

func (r *PackageCleanupPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: cleanupPolicySchemaV1,
		Version:    1,
		Description: "Provides an Artifactory Package Cleanup Policy resource. This resource enable system administrators to define and customize policies based on specific criteria for removing unused binaries from across their JFrog platform. " +
			"See [Cleanup Policies](https://jfrog.com/help/r/jfrog-platform-administration-documentation/cleanup-policies) for more details.\n\n",
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

func (r PackageCleanupPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data PackageCleanupPolicyResourceModelV1

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate project_key constraints for project-level policies
	if !data.ProjectKey.IsNull() && !data.ProjectKey.IsUnknown() && data.ProjectKey.ValueString() != "" {
		// When project_key is specified, this is a project-level policy
		projectKey := data.ProjectKey.ValueString()
		policyKey := data.Key.ValueString()

		// Check that policy key starts with project key prefix
		expectedPrefix := projectKey + "-"
		if !strings.HasPrefix(policyKey, expectedPrefix) {
			resp.Diagnostics.AddAttributeError(
				path.Root("key"),
				"Invalid Project-Level Policy Key",
				fmt.Sprintf("Project-level policy key must start with the project key prefix. Expected key to start with '%s', but got '%s'. Consider using a key like '%s<policy-name>'.", expectedPrefix, policyKey, expectedPrefix),
			)
		}

		attrs := data.SearchCriteria.Attributes()

		// Check include_all_projects - must be false for project-level policies
		if includeAllProjects, ok := attrs["include_all_projects"]; ok && !includeAllProjects.IsNull() && !includeAllProjects.IsUnknown() {
			if includeAllProjectsBool, ok := includeAllProjects.(types.Bool); ok && includeAllProjectsBool.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					path.Root("search_criteria").AtName("include_all_projects"),
					"Invalid Project-Level Policy Configuration",
					"Project-level policies (when `project_key` is specified) cannot include all projects. Set `include_all_projects` to `false`.",
				)
			}
		}

		// Check included_projects - should be empty for project-level policies
		if includedProjects, ok := attrs["included_projects"]; ok && !includedProjects.IsNull() && !includedProjects.IsUnknown() {
			if includedProjectsSet, ok := includedProjects.(types.Set); ok && len(includedProjectsSet.Elements()) > 0 {
				resp.Diagnostics.AddAttributeError(
					path.Root("search_criteria").AtName("included_projects"),
					"Invalid Project-Level Policy Configuration",
					"Project-level policies (when `project_key` is specified) should have an empty `included_projects` array `[]`. The policy will automatically apply to the specified project.",
				)
			}
		}
	}

	// Schema-level validation handles the condition validation rules
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
		policyEnablement := PackageCleanupPolicyEnablementAPIModel{
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

	// Read the created resource to get computed values
	var createdPolicy PackageCleanupPolicyAPIModel
	readResponse, readErr := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetResult(&createdPolicy).
		SetError(&jfrogErrors).
		Get(r.DocumentEndpoint)

	if readErr != nil {
		utilfw.UnableToCreateResourceError(resp, readErr.Error())
		return
	}

	if readResponse.IsError() {
		utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	resp.Diagnostics.Append(plan.fromAPIModel(ctx, createdPolicy)...)
	if resp.Diagnostics.HasError() {
		return
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
		policyEnablement := PackageCleanupPolicyEnablementAPIModel{}
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

	// Read the updated resource to get computed values
	var updatedPolicy PackageCleanupPolicyAPIModel
	readResponse, readErr := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetResult(&updatedPolicy).
		SetError(&jfrogErrors).
		Get(r.DocumentEndpoint)

	if readErr != nil {
		utilfw.UnableToUpdateResourceError(resp, readErr.Error())
		return
	}

	if readResponse.IsError() {
		utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	resp.Diagnostics.Append(plan.fromAPIModel(ctx, updatedPolicy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageCleanupPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state PackageCleanupPolicyResourceModelV1

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
func (r *PackageCleanupPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	if len(parts) > 0 && parts[0] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[0])...)
	}

	if len(parts) == 2 && parts[1] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[1])...)
	}
}
