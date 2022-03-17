package xray

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDbSyncTime(t *testing.T) {
	_, fqrn, resourceName := mkNames("db_sync-", "xray_settings")
	time := "18:45"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: dbSyncTime(resourceName, time),
				Check:  resource.TestCheckResourceAttr(fqrn, "db_sync_updates_time", time),
			},
		},
	})
}

func TestDbSyncTimeNegative(t *testing.T) {
	_, _, resourceName := mkNames("db_sync-", "xray_settings")
	var invalidTime = []string{"24:00", "24:55", "", "12:0", "string", "12pm", "9:00"}
	for _, time := range invalidTime {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviders(),
			Steps: []resource.TestStep{
				{
					Config:      dbSyncTime(resourceName, time),
					ExpectError: regexp.MustCompile("Wrong format input, expected valid hour:minutes \\(HH:mm\\) form"),
				},
			},
		})
	}
}

func dbSyncTime(resourceName string, time string) string {
	return fmt.Sprintf(`
		resource "xray_settings" "%s" {
			db_sync_updates_time = "%s"
		}
`, resourceName, time)
}
