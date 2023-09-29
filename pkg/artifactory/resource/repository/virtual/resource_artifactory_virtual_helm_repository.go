package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const HelmPackageType = "helm"

var HelmVirtualSchema = utilsdk.MergeMaps(
	BaseVirtualRepoSchema,
	RetrievalCachePeriodSecondsSchema,
	map[string]*schema.Schema{
		"use_namespaces": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories\nSee https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories. Default to 'false'",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, HelmPackageType))

func ResourceArtifactoryVirtualHelmRepository() *schema.Resource {
	type HelmVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		UseNamespaces bool `json:"useNamespaces"`
	}

	unpackHelmVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := HelmVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, HelmPackageType),
			UseNamespaces: d.GetBool("use_namespaces", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &HelmVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: HelmPackageType,
				},
			},
			UseNamespaces: false,
		}, nil
	}

	return repository.MkResourceSchema(
		HelmVirtualSchema,
		packer.Default(HelmVirtualSchema),
		unpackHelmVirtualRepository,
		constructor,
	)
}
