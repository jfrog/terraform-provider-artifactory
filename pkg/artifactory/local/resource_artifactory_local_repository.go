package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/validators"
)

var legacyLocalSchema = util.MergeSchema(map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validators.RepoKeyValidator,
	},
	"package_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: validators.RepoTypeValidator,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"notes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
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
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"debian_trivial_layout": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"checksum_policy_type": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"max_unique_tags": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"snapshot_version_behavior": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"suppress_pom_consistency_checks": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"blacked_out": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"property_sets": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true,
	},
	"archive_browsing_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"calculate_yum_metadata": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"yum_root_depth": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"docker_api_version": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"enable_file_lists_indexing": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"xray_index": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
}, CompressionFormats)

func ResourceArtifactoryLocalRepository() *schema.Resource {
	packer := util.UniversalPack(util.SchemaHasKey(legacyLocalSchema))
	return repos.MkResourceSchema(legacyLocalSchema, packer, unmarshalLocalRepository, func() interface{} {
		return &MessyLocalRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass: "local",
			},
		}
	})
}

type CommonMavenGradleLocalRepositoryParams struct {
	MaxUniqueSnapshots           int    `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots,omitempty"`
	HandleReleases               *bool  `hcl:"handle_releases" json:"handleReleases,omitempty"`
	HandleSnapshots              *bool  `hcl:"handle_snapshots" json:"handleSnapshots,omitempty"`
	SuppressPomConsistencyChecks *bool  `hcl:"suppress_pom_consistency_checks" json:"suppressPomConsistencyChecks,omitempty"`
	SnapshotVersionBehavior      string `hcl:"snapshot_version_behavior" json:"snapshotVersionBehavior,omitempty"`
	ChecksumPolicyType           string `hcl:"checksum_policy_type" json:"checksumPolicyType,omitempty"`
}

type MessyDebianLocalRepositoryParams struct {
	RepositoryBaseParams
	DebianTrivialLayout *bool `hcl:"debian_trivial_layout" json:"debianTrivialLayout,omitempty"`
}

type RpmLocalRepositoryParams struct {
	RepositoryBaseParams
	YumRootDepth            int   `hcl:"yum_root_depth" json:"yumRootDepth,omitempty"`
	CalculateYumMetadata    *bool `hcl:"calculate_yum_metadata" json:"calculateYumMetadata,omitempty"`
	EnableFileListsIndexing *bool `hcl:"enable_file_lists_indexing" json:"enableFileListsIndexing,omitempty"`
}

type MessyLocalRepo struct {
	RepositoryBaseParams
	CommonMavenGradleLocalRepositoryParams
	MessyDebianLocalRepositoryParams
	DockerLocalRepositoryParams
	RpmLocalRepositoryParams
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

func unmarshalLocalRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: data}
	repo := MessyLocalRepo{}
	repo.Rclass = "local"
	repo.Key = d.GetString("key", false)
	repo.PackageType = d.GetString("package_type", false)
	repo.Description = d.GetString("description", false)
	repo.Notes = d.GetString("notes", false)
	repo.DebianTrivialLayout = d.GetBoolRef("debian_trivial_layout", false)
	repo.IncludesPattern = d.GetString("includes_pattern", false)
	repo.ExcludesPattern = d.GetString("excludes_pattern", false)
	repo.RepoLayoutRef = d.GetString("repo_layout_ref", false)
	repo.MaxUniqueTags = d.GetInt("max_unique_tags", false)
	repo.BlackedOut = d.GetBoolRef("blacked_out", false)
	repo.CalculateYumMetadata = d.GetBoolRef("calculate_yum_metadata", false)
	repo.YumRootDepth = d.GetInt("yum_root_depth", false)
	repo.ArchiveBrowsingEnabled = d.GetBoolRef("archive_browsing_enabled", false)
	repo.DockerApiVersion = d.GetString("docker_api_verision", false)
	if repo.DockerApiVersion == "" {
		repo.DockerApiVersion = "V2" // for backward compatibility
	}
	repo.EnableFileListsIndexing = d.GetBoolRef("enable_file_lists_indexing", false)
	repo.PropertySets = d.GetSet("property_sets")
	repo.HandleReleases = d.GetBoolRef("handle_releases", false)
	repo.HandleSnapshots = d.GetBoolRef("handle_snapshots", false)
	repo.ChecksumPolicyType = d.GetString("checksum_policy_type", false)
	repo.MaxUniqueSnapshots = d.GetInt("max_unique_snapshots", false)
	repo.SnapshotVersionBehavior = d.GetString("snapshot_version_behavior", false)
	repo.SuppressPomConsistencyChecks = d.GetBoolRef("suppress_pom_consistency_checks", false)
	repo.XrayIndex = d.GetBoolRef("xray_index", false)
	repo.ForceNugetAuthentication = d.GetBool("force_nuget_authentication", false)

	return repo, repo.Key, nil
}
