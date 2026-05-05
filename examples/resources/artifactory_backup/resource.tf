resource "artifactory_backup" "backup_config_name" {
  key                      = "backup_config_name"
  enabled                  = true
  cron_exp                 = "0 0 12 * * ? *"
  retention_period_hours   = 1000
  excluded_repositories    = ["my-docker-local"]
  create_archive           = false
  exclude_new_repositories = true
  send_mail_on_error       = true
  verify_disk_space        = true
  export_mission_control   = true
}