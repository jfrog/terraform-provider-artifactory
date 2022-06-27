package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryRemoteJavaRepository(repoType string, suppressPom bool) *schema.Resource {
	var javaRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
		"fetch_jars_eagerly": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. Default value is 'false'.`,
		},
		"fetch_sources_eagerly": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested. Default value is 'false'.`,
		},
		"remote_repo_checksum_policy_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "generate-if-absent",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
				"generate-if-absent",
				"fail",
				"ignore-and-generate",
				"pass-thru",
			}, false)),
			Description: `Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Default value is 'generate-if-absent'.`,
		},
		"handle_releases": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `If set, Artifactory allows you to deploy release artifacts into this repository. Default value is 'true'.`,
		},
		"handle_snapshots": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `If set, Artifactory allows you to deploy snapshot artifacts into this repository. Default value is 'true'.`,
		},
		"suppress_pom_consistency_checks": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     suppressPom,
			Description: `By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to 'true'. Default value is 'false'.`,
		},
		"reject_invalid_jars": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal". Default value is 'false'.`,
		},
	}, repository.RepoLayoutRefSchema("remote", repoType))

	type JavaRemoteRepo struct {
		RepositoryBaseParams
		FetchJarsEagerly             bool   `json:"fetchJarsEagerly"`
		FetchSourcesEagerly          bool   `json:"fetchSourcesEagerly"`
		RemoteRepoChecksumPolicyType string `json:"remoteRepoChecksumPolicyType"`
		HandleReleases               bool   `json:"handleReleases"`
		HandleSnapshots              bool   `json:"handleSnapshots"`
		SuppressPomConsistencyChecks bool   `json:"suppressPomConsistencyChecks"`
		RejectInvalidJars            bool   `json:"rejectInvalidJars"`
	}

	var unpackJavaRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{data}
		repo := JavaRemoteRepo{
			RepositoryBaseParams:         UnpackBaseRemoteRepo(data, repoType),
			FetchJarsEagerly:             d.GetBool("fetch_jars_eagerly", false),
			FetchSourcesEagerly:          d.GetBool("fetch_sources_eagerly", false),
			RemoteRepoChecksumPolicyType: d.GetString("remote_repo_checksum_policy_type", false),
			HandleReleases:               d.GetBool("handle_releases", false),
			HandleSnapshots:              d.GetBool("handle_snapshots", false),
			SuppressPomConsistencyChecks: d.GetBool("suppress_pom_consistency_checks", false),
			RejectInvalidJars:            d.GetBool("reject_invalid_jars", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(javaRemoteSchema, packer.Default(javaRemoteSchema), unpackJavaRemoteRepo, func() interface{} {
		return &JavaRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: repoType,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}
	})
}
