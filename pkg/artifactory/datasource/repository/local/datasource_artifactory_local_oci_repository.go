// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalOciRepository() *schema.Resource {
	pkr := packer.Default(local.OCILocalSchemas[local.CurrentSchemaVersion])

	constructor := func() (interface{}, error) {
		return &local.OciLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.OCIPackageType,
				Rclass:      local.Rclass,
			},
			DockerApiVersion: "V2",
			TagRetention:     1,
			MaxUniqueTags:    0, // no limit
		}, nil
	}

	return &schema.Resource{
		Schema:      local.OCILocalSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
	}
}
