resource "artifactory_trashcan_config" "my-trash-can-config" {
  enabled               = true
  retention_period_days = 14
}
