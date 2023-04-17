package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryVirtualJavaRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   packageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	var javaSchema = util.MergeMaps(
		virtual.BaseVirtualRepoSchema,
		virtual.JavaVirtualSchema,
	)

	return &schema.Resource{
		Schema:      javaSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(javaSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}
