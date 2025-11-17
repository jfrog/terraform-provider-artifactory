package virtual

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/samber/lo"
)

// VirtualDataSourceAttributes defines the attributes for virtual repository datasources
var VirtualDataSourceAttributes = lo.Assign(
	datasource_repository.BaseDataSourceAttributes,
	map[string]schema.Attribute{},
)
