package artifactory

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBackup_full(t *testing.T) {
	const BackupTemplateFull = `
resource "artifactory_backup" "backuptest" {
	key = "backuptest"
	enabled = true
    cron_exp = "0 0 12 * * ?"
}`

	const BackupTemplateUpdate = `
resource "artifactory_local_generic_repository" "test-backup-local1" {
		  key = "test-backup-local1"
}
resource "artifactory_local_generic_repository" "test-backup-local2" {
		  key = "test-backup-local2"
}
resource "artifactory_backup" "backuptest" {
	key = "backuptest"
	enabled = false
    cron_exp = "0 0 12 * * ?"
    retention_period_hours = 1000
    excluded_repositories = [ "test-backup-local1", "test-backup-local2" ]
    depends_on = [ artifactory_local_generic_repository.test-backup-local1, artifactory_local_generic_repository.test-backup-local2 ]
}`
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccBackupDestroy("backuptest"),
		ProviderFactories: testAccProviders,

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
				),
			},
		},
	})
}

func testAccBackupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_backup."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		backups := &Backups{}

		response, err := client.R().SetResult(&backups).Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, iterBackup := range backups.BackupArr {
			if iterBackup.Key == id {
				return fmt.Errorf("error: Backup config with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
