package configuration

import (
	"context"
	"net/http"

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
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

const (
	PackageCleanupPolicyEndpointPath           = "artifactory/api/cleanup/packages/policies/{policyKey}"
	PackageCleanupPolicyEnablementEndpointPath = "artifactory/api/cleanup/packages/policies/{policyKey}/enablement"
)

func NewPackageCleanupPolicyResource() resource.Resource {
	return &PackageCleanupPolicyResource{}
}

type PackageCleanupPolicyResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type PackageCleanupPolicyResourceModel struct {
	Key               types.String `tfsdk:"key"`
	Description       types.String `tfsdk:"description"`
	CronExpression    types.String `tfsdk:"cron_expression"`
	DurationInMinutes types.Int64  `tfsdk:"duration_in_minutes"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SkipTrashcan      types.Bool   `tfsdk:"skip_trashcan"`
	SearchCriteria    types.Object `tfsdk:"search_criteria"`
}

func (r PackageCleanupPolicyResourceModel) toAPIModel(ctx context.Context, apiModel *PackageCleanupPolicyAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	attrs := r.SearchCriteria.Attributes()
	searchCriteria := PackageCleanupPolicySearchCriteriaAPIModel{
		CreatedBeforeInMonths:        attrs["created_before_in_months"].(types.Int64).ValueInt64(),
		LastDownloadedBeforeInMonths: attrs["last_downloaded_before_in_months"].(types.Int64).ValueInt64(),
	}

	diags.Append(attrs["package_types"].(types.Set).ElementsAs(ctx, &searchCriteria.PackageTypes, false)...)
	diags.Append(attrs["repos"].(types.Set).ElementsAs(ctx, &searchCriteria.Repos, false)...)
	// diags.Append(attrs["included_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedPackages, false)...)
	// diags.Append(attrs["excluded_packages"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludedPackages, false)...)

	*apiModel = PackageCleanupPolicyAPIModel{
		Key:               r.Key.ValueString(),
		Description:       r.Description.ValueString(),
		CronExpression:    r.CronExpression.ValueString(),
		DurationInMinutes: r.DurationInMinutes.ValueInt64(),
		SkipTrashcan:      r.SkipTrashcan.ValueBool(),
		SearchCriteria:    searchCriteria,
	}

	return diags
}

func (r *PackageCleanupPolicyResourceModel) fromAPIModel(ctx context.Context, apiModel PackageCleanupPolicyAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.Key = types.StringValue(apiModel.Key)
	r.Description = types.StringValue(apiModel.Description)
	r.CronExpression = types.StringValue(apiModel.CronExpression)
	r.DurationInMinutes = types.Int64Value(apiModel.DurationInMinutes)
	r.Enabled = types.BoolValue(apiModel.Enabled)
	r.SkipTrashcan = types.BoolValue(apiModel.SkipTrashcan)

	packageTypes, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.PackageTypes)
	diags.Append(ds...)

	repos, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.Repos)
	diags.Append(ds...)

	// 	includedPackages, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.IncludedPackages)
	// 	diags.Append(ds...)
	//
	// 	excludePackages, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.ExcludedPackages)
	// 	diags.Append(ds...)

	searchCriteria, ds := types.ObjectValue(
		map[string]attr.Type{
			"package_types": types.SetType{ElemType: types.StringType},
			"repos":         types.SetType{ElemType: types.StringType},
			// "included_packages":                types.SetType{ElemType: types.StringType},
			// "excluded_packages":                types.SetType{ElemType: types.StringType},
			"created_before_in_months":         types.Int64Type,
			"last_downloaded_before_in_months": types.Int64Type,
		},
		map[string]attr.Value{
			"package_types": packageTypes,
			"repos":         repos,
			// "included_packages":                includedPackages,
			// "excluded_packages":                excludePackages,
			"created_before_in_months":         types.Int64Value(apiModel.SearchCriteria.CreatedBeforeInMonths),
			"last_downloaded_before_in_months": types.Int64Value(apiModel.SearchCriteria.LastDownloadedBeforeInMonths),
		},
	)
	diags.Append(ds...)

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
	SearchCriteria    PackageCleanupPolicySearchCriteriaAPIModel `json:"searchCriteria"`
}

type PackageCleanupPolicySearchCriteriaAPIModel struct {
	PackageTypes []string `json:"packageTypes"`
	Repos        []string `json:"repos"`
	// IncludedPackages             []string `json:"includedPackages"`
	// ExcludedPackages             []string `json:"excludedPackages,omitempty"`
	CreatedBeforeInMonths        int64 `json:"createdBeforeInMonths,omitempty"`
	LastDownloadedBeforeInMonths int64 `json:"lastDownloadedBeforeInMonths,omitempty"`
}

type PackageCleanupPolicyEnablementAPIModel struct {
	Enabled bool `json:"enabled"`
}

func (r *PackageCleanupPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package_cleanup_policy"
	r.TypeName = resp.TypeName
}

func (r *PackageCleanupPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
				MarkdownDescription: "Cron expression to set a schedule for policy execution. If unset, the policy can only be triggered manually.",
			},
			"duration_in_minutes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum execution duration that the policy has to run.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enables or disabled the package cleanup policy. This allows the user to run the policy manually. If a policy has a valid cron expression, then it will be scheduled for execution based on it. If a policy is disabled, its future executions will be unscheduled.",
			},
			"skip_trashcan": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Skips the step of transferring packages to the Trash Can repository when packages are deleted. Enabling this setting results in packages being permanently deleted from Artifactory after the cleanup policy is executed.",
			},
			"search_criteria": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"package_types": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf("docker", "maven", "gradle"),
							),
						},
						MarkdownDescription: "Types of packages to be removed. Support: `docker`, `maven`, and `gradle`.",
					},
					"repos": schema.SetAttribute{
						ElementType:         types.StringType,
						Required:            true,
						MarkdownDescription: "List of repositories to clean up.",
					},
					// "included_packages": schema.SetAttribute{
					// 	ElementType: types.StringType,
					// 	Required:    true,
					// 	Validators: []validator.Set{
					// 		setvalidator.SizeBetween(1, 1),
					// 	},
					// 	MarkdownDescription: "Pattern for local repository name(s) which to be cleaned up. It accept only single element which can be specific package or pattern, and for including all packages use `**`. Example -> \"includedPackages\": [\"**\"]",
					// },
					// "excluded_packages": schema.SetAttribute{
					// 	ElementType:         types.StringType,
					// 	Optional:            true,
					// 	MarkdownDescription: "List of local repository name(s) excludes from being cleaned up. It can not accept any pattern only list of specific packages.",
					// },
					"created_before_in_months": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Remove packages based on when they were created.",
					},
					"last_downloaded_before_in_months": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Remove packages based on when they were last downloaded.",
					},
				},
				Required: true,
			},
		},
		Description: "Provides an Artifactory Package Cleanup Policy resource. This resource enable system administrators to define and customize policies based on specific criteria for removing unused binaries from across their JFrog platform.",
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

	var plan PackageCleanupPolicyResourceModel

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

	var state PackageCleanupPolicyResourceModel
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

	var plan PackageCleanupPolicyResourceModel
	var state PackageCleanupPolicyResourceModel

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

	enabledChanged := state.Enabled.ValueBool() != plan.Enabled.ValueBool()

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

	var state PackageCleanupPolicyResourceModel

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
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
