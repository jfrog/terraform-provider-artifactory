package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryHelmVirtualRepository() *schema.Resource {

	const packageType = "helm"

	helmVirtualSchema := mergeSchema(getBaseVirtualRepoSchema(packageType), map[string]*schema.Schema{
		"use_namespaces": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "(Optional) From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories\nSee https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories. Default to 'false'",
		},
	})

	type HelmVirtualRepositoryParams struct {
		VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs
		UseNamespaces bool `json:"useNamespaces"`
	}

	unpackHelmVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{data}
		repo := HelmVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: unpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, "helm"),
			UseNamespaces: d.getBool("use_namespaces", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() interface{} {
		return &HelmVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
				VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
					Rclass:      "virtual",
					PackageType: packageType,
				},
			},
			UseNamespaces: false,
		}
	}

	return mkResourceSchema(helmVirtualSchema, defaultPacker, unpackHelmVirtualRepository, constructor)
}
