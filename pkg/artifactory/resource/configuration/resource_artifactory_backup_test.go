package configuration_test

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestAccBackup_full(t *testing.T) {
	const BackupTemplateFull = `
resource "artifactory_backup" "backuptest" {
    key = "backuptest"
    enabled = true
    cron_exp = "0 0 2 ? * MON-SAT *"
}`

	const BackupTemplateUpdate = `
resource "artifactory_local_generic_repository" "test-backup-local1" {
    key = "test-backup-local1"
}

resource "artifactory_local_generic_repository" "test-backup-local2" {
    key = "test-backup-local2"
}

resource "artifactory_backup" "backuptest" {
    key                    = "backuptest"
    enabled                = false
    cron_exp               = "0 0 12 * * ? *"
    retention_period_hours = 1000
    excluded_repositories  = [ "test-backup-local1", "test-backup-local2" ]
	create_archive         = true
	verify_disk_space      = true
	export_mission_control = true

    depends_on = [ artifactory_local_generic_repository.test-backup-local1, artifactory_local_generic_repository.test-backup-local2 ]
}`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccBackupDestroy("backuptest"),

		Steps: []resource.TestStep{
			{
				Config: BackupTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "enabled", "true"),
				),
			},
			{
				Config: BackupTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "enabled", "false"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "retention_period_hours", "1000"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "excluded_repositories.#", "2"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "excluded_repositories.0", "test-backup-local1"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "excluded_repositories.1", "test-backup-local2"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "create_archive", "true"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "verify_disk_space", "true"),
					resource.TestCheckResourceAttr("artifactory_backup.backuptest", "export_mission_control", "true"),
				),
			},
		},
	})
}

func TestAccCronExpressions(t *testing.T) {
	cronExpressions := [...]string{
		"10/20 15 14 5-10 * ? *",
		"* 5,7,9 14-16 * * ? *",
		"* 5,7,9 14/2 * * WED,Sat *",
		"* * * * * ? *",
		"* * 14/2 * * mon/3 *",
		"* 5-9 14/2 * * 1-3 *",
		"*/3 */51 */12 */2 */4 ? *",
		"* 5 22-23 * * Sun *",
	}
	for _, cron := range cronExpressions {
		title := fmt.Sprintf("TestBackupCronExpression_%s", cases.Title(language.AmericanEnglish).String(cron))
		t.Run(title, func(t *testing.T) {
			resource.Test(cronTestCase(cron, t))
		})
	}
}

func cronTestCase(cronExpression string, t *testing.T) (*testing.T, resource.TestCase) {
	resourceName := fmt.Sprintf("artifactory_backup.backuptest")

	fields := map[string]interface{}{
		"cron_exp": cronExpression,
	}

	const BackupTemplateFull = `
	resource "artifactory_backup" "backuptest" {
		key = "backuptest"
		enabled = true
		cron_exp = "{{ .cron_exp }}"
	}`

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(resourceName, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("backup", BackupTemplateFull, fields),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cron_exp", cronExpression),
				),
			},
		},
	}
}

func testAccBackupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_backup."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		backups := &configuration.Backups{}

		response, err := client.R().SetResult(&backups).Get("artifactory/api/system/configuration")
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read. Response:%#v", response)
		}

		for _, iterBackup := range backups.BackupArr {
			if iterBackup.Key == id {
				return fmt.Errorf("error: Backup config with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
