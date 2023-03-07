---
subcategory: "Configuration"
---
# Artifactory Backup Resource

This resource can be used to manage the automatic and periodic backups of the entire Artifactory instance.

When an `artifactory_backup` resource is configured and enabled to true, backup of the entire Artifactory system will be done automatically and periodically.
The backup process creates a time-stamped directory in the target backup directory.

~>The `artifactory_backup` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory Backup system config
resource "artifactory_backup" "backup_config_name" {
  key                       = "backup_config_name"
  enabled                   = true
  cron_exp                  = "0 0 12 * * ? *"
  retention_period_hours    = 1000
  excluded_repositories     = []
  create_archive            = false
  exclude_new_repositories  = true
  send_mail_on_error        = true
  verify_disk_space         = true
  export_mission_control    = true
}
```
Note: `Key` argument has to match to the resource name.
Reference Link: [JFrog Artifactory Backup](https://www.jfrog.com/confluence/display/JFROG/Backups)

## Argument Reference

The following arguments are supported:

* `key`                          - (Required) The unique ID of the artifactory backup config.
* `enabled`                      - (Optional) Flag to enable or disable the backup config. Default value is `true`.
* `cron_exp`                     - (Required) A valid CRON expression that you can use to control backup frequency. Eg: "0 0 12 * * ? *", "0 0 2 ? * MON-SAT *". Note: please use 7 character format - Seconds, Minutes Hours, Day Of Month, Month, Day Of Week, Year. Also, specifying both a day-of-week AND a day-of-month parameter is not supported. One of them should be replaced by `?`. Incorrect: `* 5,7,9 14/2 * * WED,SAT *`, correct: `* 5,7,9 14/2 ? * WED,SAT *`. See details in [Cron Trigger Tutorial](http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html) and in [Cronexp package readme](https://github.com/gorhill/cronexpr#other-details).
* `retention_period_hours`       - (Optional) The number of hours to keep a backup before Artifactory will clean it up to free up disk space. Applicable only to non-incremental backups. Default value is 168 hours ie: 7 days.
* `excluded_repositories`        - (Optional) A list of excluded repositories from the backup. Default is empty list.
* `create_archive`               - (Optional) If set, backups will be created within a Zip archive (Slow and CPU intensive). Default value is `false`.
* `exclude_new_repositories`     - (Optional) When set, new repositories will not be automatically added to the backup. Default value is `false`.
* `send_mail_on_error`           - (Optional) If set, all Artifactory administrators will be notified by email if any problem is encountered during backup. Default value is `true`.
* `verify_disk_space`            - (Optional) If set, Artifactory will verify that the backup target location has enough disk space available to hold the backed up data. If there is not enough space available, Artifactory will abort the backup and write a message in the log file. Applicable only to non-incremental backups.
* `export_mission_control`       - (Optional) When set to true, mission control will not be automatically added to the backup. Default value is `false`.

## Import

Backup config can be imported using the key, e.g.

```
$ terraform import artifactory_backup.backup_name backup_name
```
