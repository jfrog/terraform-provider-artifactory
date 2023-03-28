package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func GetJavaRepoSchema(packageType string, suppressPom bool) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseLocalRepoSchema,
		map[string]*schema.Schema{
			"checksum_policy_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "client-checksums",
				ValidateDiagFunc: validator.StringInSlice(true, "client-checksums", "server-generated-checksums"),
				Description: "Checksum policy determines how Artifactory behaves when a client checksum for a deployed " +
					"resource is missing or conflicts with the locally calculated checksum (bad checksum). " +
					`Options are: "client-checksums", or "server-generated-checksums". Default: "client-checksums"\n ` +
					"For more details, please refer to Checksum Policy - " +
					"https://www.jfrog.com/confluence/display/JFROG/Local+Repositories#LocalRepositories-ChecksumPolicy",
			},
			"snapshot_version_behavior": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "unique",
				ValidateDiagFunc: validator.StringInSlice(true, "unique", "non-unique", "deployer"),
				Description: "Specifies the naming convention for Maven SNAPSHOT versions.\nThe options are " +
					"-\nunique: Version number is based on a time-stamp (default)\nnon-unique: Version number uses a" +
					" self-overriding naming pattern of artifactId-version-SNAPSHOT.type\ndeployer: Respects the settings " +
					"in the Maven client that is deploying the artifact.",
			},
			"max_unique_snapshots": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				Description: "The maximum number of unique snapshots of a single artifact to store.\nOnce the number of " +
					"snapshots exceeds this setting, older versions are removed.\nA value of 0 (default) indicates there is " +
					"no limit, and unique snapshots are not cleaned up.",
			},
			"handle_releases": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set, Artifactory allows you to deploy release artifacts into this repository.",
			},
			"handle_snapshots": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set, Artifactory allows you to deploy snapshot artifacts into this repository.",
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  suppressPom,
				Description: "By default, Artifactory keeps your repositories healthy by refusing POMs with incorrect " +
					"coordinates (path).\n  If the groupId:artifactId:version information inside the POM does not match the " +
					"deployed path, Artifactory rejects the deployment with a \"409 Conflict\" error.\n  You can disable this " +
					"behavior by setting the Suppress POM Consistency Checks checkbox.",
			},
		},
		repository.RepoLayoutRefSchema(rclass, packageType),
	)
}

type JavaLocalRepositoryParams struct {
	RepositoryBaseParams
	ChecksumPolicyType           string `hcl:"checksum_policy_type" json:"checksumPolicyType"`
	SnapshotVersionBehavior      string `hcl:"snapshot_version_behavior" json:"snapshotVersionBehavior"`
	MaxUniqueSnapshots           int    `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	HandleReleases               bool   `hcl:"handle_releases" json:"handleReleases"`
	HandleSnapshots              bool   `hcl:"handle_snapshots" json:"handleSnapshots"`
	SuppressPomConsistencyChecks bool   `hcl:"suppress_pom_consistency_checks" json:"suppressPomConsistencyChecks"`
}

var UnpackLocalJavaRepository = func(data *schema.ResourceData, rclass string, packageType string) JavaLocalRepositoryParams {
	d := &util.ResourceData{ResourceData: data}
	return JavaLocalRepositoryParams{
		RepositoryBaseParams:         UnpackBaseRepo(rclass, data, packageType),
		ChecksumPolicyType:           d.GetString("checksum_policy_type", false),
		SnapshotVersionBehavior:      d.GetString("snapshot_version_behavior", false),
		MaxUniqueSnapshots:           d.GetInt("max_unique_snapshots", false),
		HandleReleases:               d.GetBool("handle_releases", false),
		HandleSnapshots:              d.GetBool("handle_snapshots", false),
		SuppressPomConsistencyChecks: d.GetBool("suppress_pom_consistency_checks", false),
	}
}

func ResourceArtifactoryLocalJavaRepository(packageType string, suppressPom bool) *schema.Resource {
	var unPackLocalJavaRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalJavaRepository(data, rclass, packageType)
		return repo, repo.Id(), nil
	}

	javaLocalSchema := GetJavaRepoSchema(packageType, suppressPom)

	constructor := func() (interface{}, error) {
		return &JavaLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      rclass,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	return repository.MkResourceSchema(javaLocalSchema, packer.Default(javaLocalSchema), unPackLocalJavaRepository, constructor)
}
