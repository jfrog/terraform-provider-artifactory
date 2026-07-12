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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type ansibleRepo struct {
	remote.RepositoryRemoteBaseParams
}

var ansibleSchema = lo.Assign(
	remote.BaseSchema,
	map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://galaxy.ansible.com",
			Description: "The remote repo URL. Default to 'https://galaxy.ansible.com'",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.AnsiblePackageType),
)

var ansibleSchemas = remote.GetSchemas(ansibleSchema)

func DataSourceArtifactoryRemoteAnsibleRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.AnsiblePackageType)
		if err != nil {
			return nil, err
		}

		return &ansibleRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.AnsiblePackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	ansibleSchema := getSchema(ansibleSchemas)

	return &schema.Resource{
		Schema:      ansibleSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ansibleSchema), constructor),
		Description: "Provides a data source for a remote Ansible repository",
	}
}
