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

var HelmOCISchema = lo.Assign(
	remote.BaseSchema,
	map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Also known as 'Foreign Layers Caching' on the UI, default is `false`.",
		},
		"enable_token_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Enable token (Bearer) based authentication.",
		},
		// We need to set default to ["**"] once we migrate to plugin-framework. SDKv2 doesn't support that.
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "Optional include patterns to match external URLs. Ant-style path expressions are supported (*, **, ?). " +
				"For example, specifying `**/github.com/**` will only allow downloading foreign layers from github.com host." +
				"By default, this is set to '**' in the UI, which means that foreign layers may be downloaded from any external host." +
				"Due to Terraform SDKv2 limitations, we can't set the default value for the list." +
				"This value must be assigned to the attribute manually, if user don't specify any other non-default values." +
				"This attribute must be set together with `external_dependencies_enabled = true`",
		},
		"project_id": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific " +
				"project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id " +
				"will default to their default project as you have set up in your account.",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.HelmOCIPackageType),
)

var HelmOCISchemas = remote.GetSchemas(HelmOCISchema)

type HelmOciRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	ProjectId                    string   `json:"dockerProjectId"`
}

func DataSourceArtifactoryRemoteHelmOciRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.HelmOCIPackageType)
		if err != nil {
			return nil, err
		}

		return &HelmOciRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.HelmOCIPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	ociSchema := getSchema(HelmOCISchemas)

	return &schema.Resource{
		Schema:      ociSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ociSchema), constructor),
		Description: "Provides a data source for a remote Helm OCI repository",
	}
}
