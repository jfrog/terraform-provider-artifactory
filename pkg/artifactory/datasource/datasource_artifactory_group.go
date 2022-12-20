package datasource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
)

func ArtifactoryGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: security.DataSourceGroupRead,
		Schema:      security.GetDataSourceGroupSchema(),
		Description: "Provides the Artifactory Group data source. Contains information about the group configuration and optionally provides its associated user list.",
	}
}
