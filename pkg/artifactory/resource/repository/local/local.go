package local

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

const rclass = "local"

var PackageTypesLikeGeneric = []string{
	"bower",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"helm",
	"npm",
	"opkg",
	"pub",
	"puppet",
	"pypi",
	"swift",
	"terraformbackend",
	"vagrant",
}

type RepositoryBaseParams struct {
	Key                    string   `hcl:"key" json:"key,omitempty"`
	ProjectKey             string   `json:"projectKey"`
	ProjectEnvironments    []string `json:"environments"`
	Rclass                 string   `json:"rclass"`
	PackageType            string   `hcl:"package_type" json:"packageType,omitempty"`
	Description            string   `hcl:"description" json:"description,omitempty"`
	Notes                  string   `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern        string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern        string   `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
	RepoLayoutRef          string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	BlackedOut             *bool    `hcl:"blacked_out" json:"blackedOut,omitempty"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `hcl:"property_sets" json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled *bool    `hcl:"archive_browsing_enabled" json:"archiveBrowsingEnabled,omitempty"`
	DownloadRedirect       *bool    `hcl:"download_direct" json:"downloadRedirect,omitempty"`
	CdnRedirect            *bool    `json:"cdnRedirect"`
	PriorityResolution     bool     `hcl:"priority_resolution" json:"priorityResolution"`
	TerraformType          string   `json:"terraformType"`
}

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}

var BaseLocalRepoSchema = util.MergeMaps(
	repository.BaseRepoSchema,
	map[string]*schema.Schema{
		"includes_pattern": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
		},
		"excludes_pattern": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
		},
		"blacked_out": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
		},
		"xray_index": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.",
		},
		"priority_resolution": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
		},
		"property_sets": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Set:         schema.HashString,
			Optional:    true,
			Description: "List of property set name",
		},
		"archive_browsing_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
		},
		"download_direct": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
		},
		"cdn_redirect": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
		},
	},
)

// GetPackageType `packageType` in the API call payload for Terraform repositories must be "terraform", but we use
// `terraform_module` and `terraform_provider` as a package types in the Provider. GetPackageType function corrects this discrepancy.
func GetPackageType(repoType string) string {
	if strings.Contains(repoType, "terraform_") {
		return "terraform"
	}
	return repoType
}

func UnpackBaseRepo(rclassType string, s *schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &util.ResourceData{ResourceData: s}
	return RepositoryBaseParams{
		Rclass:                 rclassType,
		Key:                    d.GetString("key", false),
		ProjectKey:             d.GetString("project_key", false),
		ProjectEnvironments:    d.GetSet("project_environments"),
		PackageType:            GetPackageType(packageType),
		Description:            d.GetString("description", false),
		Notes:                  d.GetString("notes", false),
		IncludesPattern:        d.GetString("includes_pattern", false),
		ExcludesPattern:        d.GetString("excludes_pattern", false),
		RepoLayoutRef:          d.GetString("repo_layout_ref", false),
		BlackedOut:             d.GetBoolRef("blacked_out", false),
		ArchiveBrowsingEnabled: d.GetBoolRef("archive_browsing_enabled", false),
		PropertySets:           d.GetSet("property_sets"),
		XrayIndex:              d.GetBool("xray_index", false),
		DownloadRedirect:       d.GetBoolRef("download_direct", false),
		CdnRedirect:            d.GetBoolRef("cdn_redirect", false),
		PriorityResolution:     d.GetBool("priority_resolution", false),
	}
}
