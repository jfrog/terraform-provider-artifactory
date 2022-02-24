# Artifactory Backup Resource

This resource can be used to manage the automatic and periodic backups of the entire Artifactory instance.

When a artifactory_backup resource is configured and enabled to true, backup of the entire Artifactory system will be done automatically and periodically.  The backup process creates a time-stamped directory in the target backup directory.

## Example Usage

```hcl
# Configure Artifactory LDAP setting
resource "artifactory_backup" "backup_config_name" {
  key = "backup_config_name"
  enabled = true
  cron_exp = "0 0 /12 * * ?"
  retention_period_hours = 1000
  excluded_repositories = []
  create_archive = false
  exclude_new_repositories = true
  send_mail_on_error = true
}
```
Note: `Key` argument has to match to the resource name.   
Reference Link: [JFrog Artifactory Backup](https://www.jfrog.com/confluence/display/JFROG/Backups)

## Argument Reference

The following arguments are supported:

* `key`                          - (Required) The unique ID of the artifactory backup config.
* `enabled`                      - (Optional) Flag to enable or disable the backup config. Default value is `true`.
* `cron_exp`                     - (Required) A valid CRON expression that you can use to control backup frequency. Eg: "0 0 12 * * ? "
* `retention_period_hours`       - (Optional) The number of hours to keep a backup before Artifactory will clean it up to free up disk space. Applicable only to non-incremental backups. Default value is 168 hours ie: 7 days.
* `excluded_repositories`        - (Optional) A list of excluded repositories from the backup. Default is empty list.
* `create_archive`               - (Optional) If set, backups will be created within a Zip archive (Slow and CPU intensive). Default value is `false`.
* `exclude_new_repositories`     - (Optional) When set, new repositories will not be automatically added to the backup. Default value is `false`.
* `send_mail_on_error`           - (Optional) If set, all Artifactory administrators will be notified by email if any problem is encountered during backup. Default value is `true`.

## Import

Backup config can be imported using the key, e.g.

```
$ terraform import artifactory_backup.backup_name backup_name
```
