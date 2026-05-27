resource "artifactory_item_properties" "my-repo-properties" {
  repo_key = "my-generic-local"
  properties = {
    "key1" : ["value1"],
    "key2" : ["value2", "value3"]
  }
  is_recursive = true
}

resource "artifactory_item_properties" "my-folder-properties" {
  repo_key  = "my-generic-local"
  item_path = "folder/subfolder"
  properties = {
    "key1" : ["value1"],
    "key2" : ["value2", "value3"]
  }
  is_recursive = true
}