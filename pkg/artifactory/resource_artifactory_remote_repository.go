package artifactory

import (
	"fmt"

	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var legacyRemoteSchema = mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"handle_releases": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"handle_snapshots": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"max_unique_snapshots": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"suppress_pom_consistency_checks": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"remote_repo_checksum_policy_type": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.StringInSlice([]string{
			"generate-if-absent",
			"fail",
			"ignore-and-generate",
			"pass-thru",
		}, false),
	},
	"fetch_jars_eagerly": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"fetch_sources_eagerly": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"pypi_registry_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"bower_registry_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"enable_token_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"vcs_type": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"vcs_git_provider": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"vcs_git_download_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"feed_context_path": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"download_context_path": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"v3_feed_url": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"propagate_query_params": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
		DefaultFunc: func() (interface{}, error) {
			return false, nil
		},
	},
})

type MessyRemoteRepo struct {
	services.RemoteRepositoryBaseParams
	services.BowerRemoteRepositoryParams
	services.CommonMavenGradleRemoteRepositoryParams
	services.DockerRemoteRepositoryParams
	services.VcsRemoteRepositoryParams
	services.PypiRemoteRepositoryParams
	services.NugetRemoteRepositoryParams
	PropagateQueryParams bool `json:"propagateQueryParams"`
}

var legacyRemoteRepoReadFun = mkRepoRead(packLegacyRemoteRepo, func() interface{} {
	return &MessyRemoteRepo{}
})

func resourceArtifactoryRemoteRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackLegacyRemoteRepo, legacyRemoteRepoReadFun),
		Read:   legacyRemoteRepoReadFun,
		Update: mkRepoUpdate(unpackLegacyRemoteRepo, legacyRemoteRepoReadFun),
		Delete: deleteRepo,
		Exists: repoExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: legacyRemoteSchema,
		DeprecationMessage: "This generic-ish repo type is deprecated and the repo specific resource should be used instead",
	}
}

func unpackLegacyRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := MessyRemoteRepo{
		RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s),
	}

	repo.RemoteRepoChecksumPolicyType = d.getString("remote_repo_checksum_policy_type", true)
	repo.BowerRegistryUrl = d.getString("bower_registry_url", true)
	repo.EnableTokenAuthentication = d.getBoolRef("enable_token_authentication", true)
	repo.FetchJarsEagerly = d.getBoolRef("fetch_jars_eagerly", true)
	repo.FetchSourcesEagerly = d.getBoolRef("fetch_sources_eagerly", true)
	repo.HandleReleases = d.getBoolRef("handle_releases", true)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", true)
	repo.MaxUniqueSnapshots = d.getInt("max_unique_snapshots", true)
	repo.PypiRegistryUrl = d.getString("pypi_registry_url", true)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", true)
	repo.VcsGitDownloadUrl = d.getString("vcs_git_download_url", true)
	repo.VcsGitProvider = d.getString("vcs_git_provider", true)
	repo.VcsType = d.getString("vcs_type", true)
	repo.FeedContextPath = d.getString("feed_context_path", true)
	repo.DownloadContextPath = d.getString("download_context_path", true)
	repo.V3FeedUrl = d.getString("v3_feed_url", true)
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)
	repo.PropagateQueryParams = d.getBool("propagate_query_params", true)

	if repo.PackageType != "" && repo.PackageType != "generic" && repo.PropagateQueryParams == true {
		return MessyRemoteRepo{}, "", fmt.Errorf("cannot use propagate_query_params with repository type %s. This parameter can be used only with generic repositories", repo.PackageType)
	}
	return repo, repo.Key, nil
}

func packLegacyRemoteRepo(r interface{}, d *schema.ResourceData) error {
	repo := r.(*MessyRemoteRepo)
	setValue := packBaseRemoteRepo(d, repo.RemoteRepositoryBaseParams)

	setValue("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType)
	setValue("bower_registry_url", repo.BowerRegistryUrl)
	setValue("enable_token_authentication", repo.EnableTokenAuthentication)
	setValue("fetch_jars_eagerly", repo.FetchJarsEagerly)
	setValue("fetch_sources_eagerly", repo.FetchSourcesEagerly)
	setValue("handle_releases", repo.HandleReleases)
	setValue("handle_snapshots", repo.HandleSnapshots)
	setValue("max_unique_snapshots", repo.MaxUniqueSnapshots)
	setValue("pypi_registry_url", repo.PypiRegistryUrl)
	setValue("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks)
	setValue("vcs_git_download_url", repo.VcsGitDownloadUrl)
	setValue("vcs_git_provider", repo.VcsGitProvider)
	setValue("vcs_type", repo.VcsType)
	setValue("feed_context_path", repo.FeedContextPath)
	setValue("download_context_path", repo.DownloadContextPath)
	setValue("v3_feed_url", repo.V3FeedUrl)
	setValue("force_nuget_authentication", repo.ForceNugetAuthentication)
	errors := setValue("propagate_query_params", repo.PropagateQueryParams)

	if repo.Password != "" {
		errors = setValue("password", getMD5Hash(repo.Password))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack remote repo %q", errors)
	}
	return nil
}
