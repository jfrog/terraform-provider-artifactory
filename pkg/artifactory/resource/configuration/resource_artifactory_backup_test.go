package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccBackup_full(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("backup-", "artifactory_backup")
	_, _, repoResourceName1 := testutil.MkNames("test-backup-local-", "artifactory_local_generic_repository")
	_, _, repoResourceName2 := testutil.MkNames("test-backup-local-", "artifactory_local_generic_repository")

	const BackupTemplateFull = `
resource "artifactory_backup" "{{ .resourceName }}" {
    key = "{{ .resourceName }}"
    enabled = true
    cron_exp = "0 0 2 ? * MON-SAT *"
}`

	testData := map[string]string{
		"resourceName":      resourceName,
		"repoResourceName1": repoResourceName1,
		"repoResourceName2": repoResourceName2,
	}

	const BackupTemplateUpdate = `
resource "artifactory_local_generic_repository" "{{ .repoResourceName1 }}" {
    key = "{{ .repoResourceName1 }}"
}

resource "artifactory_local_generic_repository" "{{ .repoResourceName2 }}" {
    key = "{{ .repoResourceName2 }}"
}

resource "artifactory_backup" "{{ .resourceName }}" {
    key                    = "{{ .resourceName }}"
    enabled                = true
    cron_exp               = "0 0 12 * * ? *"
    retention_period_hours = 1000
    excluded_repositories  = [
		artifactory_local_generic_repository.{{ .repoResourceName1 }}.key,
		artifactory_local_generic_repository.{{ .repoResourceName2 }}.key,
	]
	create_archive         = true
	verify_disk_space      = true
	export_mission_control = true
}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy:             testAccBackupDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, BackupTemplateFull, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 2 ? * MON-SAT *"),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate(fqrn, BackupTemplateUpdate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 12 * * ? *"),
					resource.TestCheckResourceAttr(fqrn, "retention_period_hours", "1000"),
					resource.TestCheckResourceAttr(fqrn, "excluded_repositories.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "excluded_repositories.0", repoResourceName1),
					resource.TestCheckResourceAttr(fqrn, "excluded_repositories.1", repoResourceName2),
					resource.TestCheckResourceAttr(fqrn, "create_archive", "true"),
					resource.TestCheckResourceAttr(fqrn, "verify_disk_space", "true"),
					resource.TestCheckResourceAttr(fqrn, "export_mission_control", "true"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportStateId:                        resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccBackup_importNotFound(t *testing.T) {
	config := `
		resource "artifactory_backup" "not-exist-test" {
		  enabled                = false
		  cron_exp               = "0 0 12 * * ? *"
		  retention_period_hours = 1000
		  excluded_repositories  = []
		  create_archive         = true
		  verify_disk_space      = true
		  export_mission_control = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_backup.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func TestAccBackup_invalid_cron(t *testing.T) {
	config := `
		resource "artifactory_backup" "invalid-cron-test" {
		  key                    = "invalid-cron-test"
		  enabled                = false
		  cron_exp               = "foo"
		  retention_period_hours = 1000
		  excluded_repositories  = []
		  create_archive         = true
		  verify_disk_space      = true
		  export_mission_control = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:       config,
				ResourceName: "artifactory_backup.invalid-cron-test",
				ExpectError:  regexp.MustCompile("value must be a valid cron expression"),
			},
		},
	})
}

func TestAccBackup_CronExpressions(t *testing.T) {
	cronExpressions := [...]string{
		"10/20 15 14 5-10 * ? *",
		"* 5,7,9 14-16 * * ? *",
		"* 5,7,9 14/2 * * WED,Sat *",
		"* * * * * ? *",
		"* * 14/2 * * mon/3 *",
		"* 5-9 14/2 * * 1-3 *",
		"*/3 */51 */12 */2 */4 ? *",
		"* 5 22-23 * * Sun *",
		"0/5 14,18,3-39,52 * ? JAN,MAR,SEP MON-FRI 2002-2010",
	}
	for _, cron := range cronExpressions {
		t.Run(cron, func(t *testing.T) {
			resource.Test(cronTestCase(cron, t))
		})
	}
}

func cronTestCase(cronExpression string, t *testing.T) (*testing.T, resource.TestCase) {
	fqrn := "artifactory_backup.backuptest"

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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate("backup", BackupTemplateFull, fields),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "cron_exp", cronExpression),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportStateId:                        "backuptest",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	}
}

func testAccBackupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

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
