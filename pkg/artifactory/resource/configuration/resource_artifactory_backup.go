package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
	"gopkg.in/yaml.v3"
)

type Backup struct {
	Key                    string   `xml:"key" yaml:"key"`
	CronExp                string   `xml:"cronExp" yaml:"cronExp"`
	Enabled                bool     `xml:"enabled" yaml:"enabled"`
	RetentionPeriodHours   int      `xml:"retentionPeriodHours" yaml:"retentionPeriodHours"`
	ExcludedRepositories   []string `xml:"excludedRepositories>repositoryRef" yaml:"excludedRepositories"`
	CreateArchive          bool     `xml:"createArchive" yaml:"createArchive"`
	ExcludeNewRepositories bool     `xml:"excludeNewRepositories" yaml:"excludeNewRepositories"`
	SendMailOnError        bool     `xml:"sendMailOnError" yaml:"sendMailOnError"`
	VerifyDiskSpace        bool     `xml:"precalculate" yaml:"precalculate"`
	ExportMissionControl   bool     `xml:"exportMissionControl" yaml:"exportMissionControl"`
}

func (b Backup) Id() string {
	return b.Key
}

type Backups struct {
	BackupArr []Backup `xml:"backups>backup" yaml:"backup"`
}

func ResourceArtifactoryBackup() *schema.Resource {
	var backupSchema = map[string]*schema.Schema{
		"key": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "Backup config name.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Flag to enable or disable the backup config. Default value is 'true'.",
		},
		"cron_exp": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validator.Cron,
			Description:      "Cron expression to control the backup frequency.",
		},
		"retention_period_hours": {
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          168,
			ValidateDiagFunc: validator.IntAtLeast(0),
			Description:      "The number of hours to keep a backup before Artifactory will clean it up to free up disk space. Applicable only to non-incremental backups. Default value is 168 hours ie: 7 days.",
		},
		"excluded_repositories": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of excluded repositories from the backup. Default is empty list.",
		},
		"create_archive": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If set to true, backups will be created within a Zip archive (Slow and CPU intensive). Default value is 'false'",
		},
		"exclude_new_repositories": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set to true, new repositories will not be automatically added to the backup. Default value is 'false'.",
		},
		"send_mail_on_error": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "If set to true, all Artifactory administrators will be notified by email if any problem is encountered during backup. Default value is 'true'.",
		},
		"verify_disk_space": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If set, Artifactory will verify that the backup target location has enough disk space available to hold the backed up data. If there is not enough space available, Artifactory will abort the backup and write a message in the log file. Applicable only to non-incremental backups.",
		},
		"export_mission_control": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set to true, mission control will not be automatically added to the backup. Default value is 'false'.",
		},
	}

	var unpackBackup = func(s *schema.ResourceData) Backup {
		d := &util.ResourceData{ResourceData: s}
		backup := Backup{
			Key:                    d.GetString("key", false),
			Enabled:                d.GetBool("enabled", false),
			CronExp:                d.GetString("cron_exp", false),
			RetentionPeriodHours:   d.GetInt("retention_period_hours", false),
			CreateArchive:          d.GetBool("create_archive", false),
			ExcludeNewRepositories: d.GetBool("exclude_new_repositories", false),
			SendMailOnError:        d.GetBool("send_mail_on_error", false),
			ExcludedRepositories:   d.GetList("excluded_repositories"),
			VerifyDiskSpace:        d.GetBool("verify_disk_space", false),
			ExportMissionControl:   d.GetBool("export_mission_control", false),
		}
		return backup
	}

	var resourceBackupRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		data := &util.ResourceData{ResourceData: d}
		key := data.GetString("key", false)

		backups := Backups{}
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&backups).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedBackup := FindConfigurationById[Backup](backups.BackupArr, key)
		if matchedBackup == nil {
			d.SetId("")
			return nil
		}

		pkr := packer.Default(backupSchema)

		return diag.FromErr(pkr(matchedBackup, d))
	}

	var resourceBackupUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedBackup := unpackBackup(d)

		/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.

		There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.

		GET call structure has "backups -> backup -> Array of backup config blocks".

		PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".

		Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.

		See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration for patching system configuration
		using YAML
		*/
		var constructBody = map[string]map[string]Backup{}
		constructBody["backups"] = map[string]Backup{}
		constructBody["backups"][unpackedBackup.Key] = unpackedBackup
		content, err := yaml.Marshal(&constructBody)

		if err != nil {
			return diag.FromErr(err)
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.FromErr(err)
		}

		// we should only have one backup config resource, using same id
		d.SetId(unpackedBackup.Key)
		return resourceBackupRead(ctx, d, m)
	}

	var resourceBackupDelete = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		rsrcBackup := unpackBackup(d)

		deleteBackupConfig := fmt.Sprintf(`
backups:
  %s: ~
`, rsrcBackup.Key)

		err := SendConfigurationPatch([]byte(deleteBackupConfig), m)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId("")
		return nil
	}

	return &schema.Resource{
		UpdateContext: resourceBackupUpdate,
		CreateContext: resourceBackupUpdate,
		DeleteContext: resourceBackupDelete,
		ReadContext:   resourceBackupRead,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("key", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema:      backupSchema,
		Description: "Provides an Artifactory backup config resource. This resource configuration corresponds to backup config block in system configuration XML (REST endpoint: artifactory/api/system/configuration). Manages the automatic and periodic backups of the entire Artifactory instance",
	}
}
