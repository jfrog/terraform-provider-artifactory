package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

const Rclass = "virtual"

type RepositoryBaseParams struct {
	Key                                           string   `hcl:"key" json:"key,omitempty"`
	ProjectKey                                    string   `json:"projectKey"`
	ProjectEnvironments                           []string `json:"environments"`
	Rclass                                        string   `json:"rclass"`
	PackageType                                   string   `hcl:"package_type" json:"packageType,omitempty"`
	Description                                   string   `json:"description"`
	Notes                                         string   `json:"notes"`
	IncludesPattern                               string   `json:"includesPattern"`
	ExcludesPattern                               string   `json:"excludesPattern"`
	RepoLayoutRef                                 string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	Repositories                                  []string `hcl:"repositories" json:"repositories,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `hcl:"artifactory_requests_can_retrieve_remote_artifacts" json:"artifactoryRequestsCanRetrieveRemoteArtifacts"`
	DefaultDeploymentRepo                         string   `hcl:"default_deployment_repo" json:"defaultDeploymentRepo,omitempty"`
}

type RepositoryBaseParamsWithRetrievalCachePeriodSecs struct {
	RepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
}

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}

var PackageTypesLikeGeneric = []string{
	"gems",
	"generic",
	"gitlfs",
	"composer",
	"p2",
	"pub",
	"puppet",
	"pypi",
	"swift",
	"terraform",
}

var PackageTypesLikeGenericWithRetrievalCachePeriodSecs = []string{
	"chef",
	"conan",
	"conda",
	"cran",
}

var BaseVirtualRepoSchema = util.MergeMaps(
	repository.BaseRepoSchema,
	map[string]*schema.Schema{
		"repositories": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "The effective list of actual repositories included in this virtual repository.",
		},
		"artifactory_requests_can_retrieve_remote_artifacts": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.",
		},
		"default_deployment_repo": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Default repository to deploy artifacts.",
		},
	},
)

func UnpackBaseVirtRepo(s *schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &util.ResourceData{ResourceData: s}

	return RepositoryBaseParams{
		Key:                 d.GetString("key", false),
		Rclass:              Rclass,
		ProjectKey:          d.GetString("project_key", false),
		ProjectEnvironments: d.GetSet("project_environments"),
		PackageType:         packageType, // must be set independently
		IncludesPattern:     d.GetString("includes_pattern", false),
		ExcludesPattern:     d.GetString("excludes_pattern", false),
		RepoLayoutRef:       d.GetString("repo_layout_ref", false),
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: d.GetBool("artifactory_requests_can_retrieve_remote_artifacts", false),
		Repositories:          d.GetList("repositories"),
		Description:           d.GetString("description", false),
		Notes:                 d.GetString("notes", false),
		DefaultDeploymentRepo: repository.HandleResetWithNonExistentValue(d, "default_deployment_repo"),
	}
}

func UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s *schema.ResourceData, packageType string) RepositoryBaseParamsWithRetrievalCachePeriodSecs {
	d := &util.ResourceData{ResourceData: s}

	return RepositoryBaseParamsWithRetrievalCachePeriodSecs{
		RepositoryBaseParams:            UnpackBaseVirtRepo(s, packageType),
		VirtualRetrievalCachePeriodSecs: d.GetInt("retrieval_cache_period_seconds", false),
	}
}

var externalDependenciesSchema = map[string]*schema.Schema{
	"external_dependencies_enabled": {
		Type:        schema.TypeBool,
		Default:     false,
		Optional:    true,
		Description: "When set, external dependencies are rewritten. Default value is false.",
	},
	"external_dependencies_remote_repo": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		RequiredWith:     []string{"external_dependencies_enabled"},
		Description:      "The remote repository aggregated by this virtual repository in which the external dependency will be cached.",
	},
	"external_dependencies_patterns": {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		RequiredWith: []string{"external_dependencies_enabled"},
		Description: "An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. " +
			"By default, this is set to ** which means that dependencies may be downloaded from any external source.",
	},
}

type ExternalDependenciesVirtualRepositoryParams struct {
	RepositoryBaseParams
	ExternalDependenciesEnabled    bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesRemoteRepo string   `json:"externalDependenciesRemoteRepo"`
	ExternalDependenciesPatterns   []string `json:"externalDependenciesPatterns"`
}

var unpackExternalDependenciesVirtualRepository = func(s *schema.ResourceData, packageType string) ExternalDependenciesVirtualRepositoryParams {
	d := &util.ResourceData{ResourceData: s}

	return ExternalDependenciesVirtualRepositoryParams{
		RepositoryBaseParams:           UnpackBaseVirtRepo(s, packageType),
		ExternalDependenciesEnabled:    d.GetBool("external_dependencies_enabled", false),
		ExternalDependenciesRemoteRepo: d.GetString("external_dependencies_remote_repo", false),
		ExternalDependenciesPatterns:   d.GetList("external_dependencies_patterns"),
	}
}

var RetrievalCachePeriodSecondsSchema = map[string]*schema.Schema{
	"retrieval_cache_period_seconds": {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  7200,
		Description: "This value refers to the number of seconds to cache metadata files before checking for newer " +
			"versions on aggregated repositories. A value of 0 indicates no caching.",
		ValidateFunc: validation.IntAtLeast(0),
	},
}
