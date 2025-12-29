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

package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var helmSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	map[string]*schema.Schema{
		"use_namespaces": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories\nSee https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories. Default to 'false'",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.HelmPackageType),
)

var HelmSchemas = GetSchemas(helmSchema)

func ResourceArtifactoryVirtualHelmRepository() *schema.Resource {
	type HelmVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		UseNamespaces bool `json:"useNamespaces"`
	}

	unpackHelmVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := HelmVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, repository.HelmPackageType),
			UseNamespaces: d.GetBool("use_namespaces", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &HelmVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.HelmPackageType,
				},
			},
			UseNamespaces: false,
		}, nil
	}

	return repository.MkResourceSchema(
		HelmSchemas,
		packer.Default(HelmSchemas[CurrentSchemaVersion]),
		unpackHelmVirtualRepository,
		constructor,
	)
}
