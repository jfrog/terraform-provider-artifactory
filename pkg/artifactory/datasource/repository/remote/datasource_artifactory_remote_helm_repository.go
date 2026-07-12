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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

var HelmSchema = lo.Assign(
	remote.BaseSchema,
	map[string]*schema.Schema{
		"helm_charts_base_url": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.Any(
					validation.IsURLWithScheme([]string{"http", "https", "oci"}),
					validation.StringIsEmpty,
				),
			),
			Description: "Base URL for the translation of chart source URLs in the index.yaml of virtual repos. " +
				"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url. " +
				"Support http/https/oci protocol scheme.",
		},
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Default:     false,
			Optional:    true,
			Description: "When set, external dependencies are rewritten. External Dependency Rewrite in the UI.",
		},
		// We need to set default to ["**"] once we migrate to plugin-framework. SDKv2 doesn't support that.
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response." +
				"Default value in UI is empty. This attribute must be set together with `external_dependencies_enabled = true`",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.HelmPackageType),
)

var HelmSchemas = remote.GetSchemas(HelmSchema)

type HelmRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl"`
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns"`
}

func DataSourceArtifactoryRemoteHelmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.HelmPackageType)
		if err != nil {
			return nil, err
		}

		return &HelmRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.HelmPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	helmSchema := getSchema(HelmSchemas)

	return &schema.Resource{
		Schema:      helmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(helmSchema), constructor),
		Description: "Provides a data source for a remote Helm repository",
	}
}
