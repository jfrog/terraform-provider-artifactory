package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"

	"gopkg.in/yaml.v3"
)

type BackupAPIModel struct {
	Key                    string    `xml:"key" yaml:"key"`
	CronExp                string    `xml:"cronExp" yaml:"cronExp"`
	Enabled                bool      `xml:"enabled" yaml:"enabled"`
	RetentionPeriodHours   int64     `xml:"retentionPeriodHours" yaml:"retentionPeriodHours"`
	ExcludedRepositories   *[]string `xml:"excludedRepositories>repositoryRef" yaml:"excludedRepositories"`
	CreateArchive          bool      `xml:"createArchive" yaml:"createArchive"`
	ExcludeNewRepositories bool      `xml:"excludeNewRepositories" yaml:"excludeNewRepositories"`
	SendMailOnError        bool      `xml:"sendMailOnError" yaml:"sendMailOnError"`
	VerifyDiskSpace        bool      `xml:"precalculate" yaml:"precalculate"`
	ExportMissionControl   bool      `xml:"exportMissionControl" yaml:"exportMissionControl"`
}

func (m BackupAPIModel) Id() string {
	return m.Key
}

type Backups struct {
	BackupArr []BackupAPIModel `xml:"backups>backup" yaml:"backup"`
}

type BackupResourceModel struct {
	Key                    types.String `tfsdk:"key"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	CronExp                types.String `tfsdk:"cron_exp"`
	RetentionPeriodHours   types.Int64  `tfsdk:"retention_period_hours"`
	ExcludedRepositories   types.List   `tfsdk:"excluded_repositories"`
	CreateArchive          types.Bool   `tfsdk:"create_archive"`
	ExcludeNewRepositories types.Bool   `tfsdk:"exclude_new_repositories"`
	SendMailOnError        types.Bool   `tfsdk:"send_mail_on_error"`
	VerifyDiskSpace        types.Bool   `tfsdk:"verify_disk_space"`
	ExportMissionControl   types.Bool   `tfsdk:"export_mission_control"`
}

func (r *BackupResourceModel) toAPIModel(ctx context.Context, backup *BackupAPIModel) diag.Diagnostics {
	// Convert from Terraform resource model into API model
	var excludedRepositories []string
	diags := r.ExcludedRepositories.ElementsAs(ctx, &excludedRepositories, true)
	if diags != nil {
		return diags
	}

	*backup = BackupAPIModel{
		Key:                    r.Key.ValueString(),
		Enabled:                r.Enabled.ValueBool(),
		CronExp:                r.CronExp.ValueString(),
		RetentionPeriodHours:   r.RetentionPeriodHours.ValueInt64(),
		CreateArchive:          r.CreateArchive.ValueBool(),
		ExcludeNewRepositories: r.ExcludeNewRepositories.ValueBool(),
		SendMailOnError:        r.SendMailOnError.ValueBool(),
		ExcludedRepositories:   &excludedRepositories,
		VerifyDiskSpace:        r.VerifyDiskSpace.ValueBool(),
		ExportMissionControl:   r.ExportMissionControl.ValueBool(),
	}

	return nil
}

func (r *BackupResourceModel) FromAPIModel(ctx context.Context, backup *BackupAPIModel) diag.Diagnostics {
	r.Key = types.StringValue(backup.Key)
	r.Enabled = types.BoolValue(backup.Enabled)
	r.CronExp = types.StringValue(backup.CronExp)
	r.RetentionPeriodHours = types.Int64Value(backup.RetentionPeriodHours)
	r.CreateArchive = types.BoolValue(backup.CreateArchive)
	r.ExcludeNewRepositories = types.BoolValue(backup.ExcludeNewRepositories)
	r.SendMailOnError = types.BoolValue(backup.SendMailOnError)

	excludedRepositories, diags := types.ListValueFrom(ctx, types.StringType, backup.ExcludedRepositories)
	if diags != nil {
		return diags
	}
	r.ExcludedRepositories = excludedRepositories
	r.VerifyDiskSpace = types.BoolValue(backup.VerifyDiskSpace)
	r.ExportMissionControl = types.BoolValue(backup.ExportMissionControl)

	return nil
}

func NewBackupResource() resource.Resource {
	return &BackupResource{}
}

type BackupResource struct {
	ProviderData utilsdk.ProvderMetadata
}

func (r *BackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_backup"
}

func (r *BackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory backup config resource. This resource configuration corresponds to backup config block in system configuration XML (REST endpoint: artifactory/api/system/configuration). Manages the automatic and periodic backups of the entire Artifactory instance.",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Flag to enable or disable the backup config. Default value is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"cron_exp": schema.StringAttribute{
				MarkdownDescription: "Cron expression to control the backup frequency.",
				Required:            true,
				Validators: []validator.String{
					validatorfw_string.IsCron(),
				},
			},
			"retention_period_hours": schema.Int64Attribute{
				MarkdownDescription: "The number of hours to keep a backup before Artifactory will clean it up to free up disk space. Applicable only to non-incremental backups. Default value is 168 hours i.e. 7 days.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(168),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"excluded_repositories": schema.ListAttribute{
				MarkdownDescription: "List of excluded repositories from the backup. Default is empty list.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"create_archive": schema.BoolAttribute{
				MarkdownDescription: "If set to true, backups will be created within a Zip archive (Slow and CPU intensive). Default value is `false`",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"exclude_new_repositories": schema.BoolAttribute{
				MarkdownDescription: "When set to true, new repositories will not be automatically added to the backup. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"send_mail_on_error": schema.BoolAttribute{
				MarkdownDescription: "If set to true, all Artifactory administrators will be notified by email if any problem is encountered during backup. Default value is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"verify_disk_space": schema.BoolAttribute{
				MarkdownDescription: "If set, Artifactory will verify that the backup target location has enough disk space available to hold the backed up data. If there is not enough space available, Artifactory will abort the backup and write a message in the log file. Applicable only to non-incremental backups. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"export_mission_control": schema.BoolAttribute{
				MarkdownDescription: "When set to true, mission control will not be automatically added to the backup. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *BackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *BackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *BackupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var backup BackupAPIModel
	resp.Diagnostics.Append(data.toAPIModel(ctx, &backup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.

	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.

	GET call structure has "backups -> backup -> Array of backup config blocks".

	PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".

	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.

	See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration for patching system configuration
	using YAML
	*/
	var constructBody = map[string]map[string]BackupAPIModel{}
	constructBody["backups"] = map[string]BackupAPIModel{}
	constructBody["backups"][backup.Key] = backup
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Assign the resource ID for the resource in the state
	data.Key = types.StringValue(backup.Key)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *BackupResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var backups Backups
	_, err := r.ProviderData.Client.R().
		SetResult(&backups).
		Get("artifactory/api/system/configuration")
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, "failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		return
	}

	matchedBackup := FindConfigurationById[BackupAPIModel](backups.BackupArr, state.Key.ValueString())
	if matchedBackup == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("key"),
			"no matching backup found",
			state.Key.ValueString(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, matchedBackup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *BackupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	var backup BackupAPIModel
	resp.Diagnostics.Append(data.toAPIModel(ctx, &backup)...)

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.

	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.

	GET call structure has "backups -> backup -> Array of backup config blocks".

	PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".

	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.

	See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration for patching system configuration
	using YAML
	*/
	var constructBody = map[string]map[string]BackupAPIModel{}
	constructBody["backups"] = map[string]BackupAPIModel{}
	constructBody["backups"][backup.Key] = backup
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(data.FromAPIModel(ctx, &backup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBackupConfig := fmt.Sprintf(`
backups:
  %s: ~
`, data.Key.ValueString())

	err := SendConfigurationPatch([]byte(deleteBackupConfig), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
