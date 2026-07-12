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

package remote

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

var basicSchema = func(packageType string) map[string]*schema.Schema {
	return lo.Assign(
		remote.BaseSchema,
		resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, packageType),
	)
}

func DataSourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, packageType)
		if err != nil {
			return nil, err
		}

		return &remote.RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        remote.Rclass,
			RepoLayoutRef: repoLayout,
		}, nil
	}

	basicSchemas := remote.GetSchemas(basicSchema(packageType))
	basicSchema := getSchema(basicSchemas)

	return &schema.Resource{
		Schema:      basicSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(basicSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a remote %s repository", packageType),
	}
}
