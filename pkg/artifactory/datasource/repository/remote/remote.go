package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
)

var getSchema = func(schemas map[int16]map[string]*schema.Schema) map[string]*schema.Schema {
	s := schemas[remote.CurrentSchemaVersion]

	s["url"].Required = false
	s["url"].Optional = true

	return s
}
