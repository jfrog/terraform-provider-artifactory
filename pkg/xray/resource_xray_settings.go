package xray

import (
	"context"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceXrayDbSyncTime() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceXrayDbSyncTimeCreate,
		ReadContext:   resourceXrayDbSyncTimeRead,
		UpdateContext: resourceXrayDbSyncTimeUpdate,
		DeleteContext: resourceXrayDbSyncTimeDelete,
		Description:   "Provides an Xray DB Sync Time resource.",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"db_sync_updates_time": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The time of the DB sync daily update job.",
				ValidateDiagFunc: matchesHoursMinutesTime,
			},
		},
	}
}

type DbSyncDailyUpdatesTime struct {
	DbSyncTime string `json:"db_sync_updates_time"`
}

func unpackDBSyncTime(s *schema.ResourceData) DbSyncDailyUpdatesTime {
	d := &ResourceData{s}
	dbSyncTime := DbSyncDailyUpdatesTime{
		DbSyncTime: d.getString("db_sync_updates_time", false),
	}
	return dbSyncTime
}

func packDBSyncTime(dbSyncTime DbSyncDailyUpdatesTime, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("db_sync_updates_time", dbSyncTime.DbSyncTime); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getDbSyncTime(client *resty.Client) (DbSyncDailyUpdatesTime, *resty.Response, error) {
	dbSyncTime := DbSyncDailyUpdatesTime{}
	resp, err := client.R().SetResult(&dbSyncTime).Get("xray/api/v1/configuration/dbsync/time")
	return dbSyncTime, resp, err
}

// No create functionality provided by API. DB sync time exists by default and can only be modified
func resourceXrayDbSyncTimeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceXrayDbSyncTimeUpdate(ctx, d, m)
}

func resourceXrayDbSyncTimeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	dbSyncTime, resp, err := getDbSyncTime(m.(*resty.Client))
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] DB sync settings (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return diag.FromErr(err)
	}
	packDBSyncTime(dbSyncTime, d)
	return nil
}

func resourceXrayDbSyncTimeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	dbSyncTime := unpackDBSyncTime(d)
	resp, err := m.(*resty.Client).R().SetBody(dbSyncTime).Put("xray/api/v1/configuration/dbsync/time")
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] DB sync settings (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return diag.FromErr(err)
	}

	d.SetId(dbSyncTime.DbSyncTime)
	return resourceXrayDbSyncTimeRead(ctx, d, m)
}

// No delete functionality provided by API for the DB sync call.
// Delete function will remove the object from the Terraform state
func resourceXrayDbSyncTimeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
