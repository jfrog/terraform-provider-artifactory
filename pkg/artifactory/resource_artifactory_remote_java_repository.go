package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceArtifactoryRemoteJavaRepository(repoType string, suppressPom bool) *schema.Resource {
	var javaRemoteSchema = mergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"fetch_jars_eagerly": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `(Optional) When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. Default value is 'false'.`,
		},
		"fetch_sources_eagerly": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `(Optional) When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested. Default value is 'false'.`,
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
			Description: `(Optional) Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Default value is 'generate-if-absent'.`,
		},
		"handle_releases": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `(Optional) If set, Artifactory allows you to deploy release artifacts into this repository. Default value is 'true'.`,
		},
		"handle_snapshots": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `(Optional) If set, Artifactory allows you to deploy snapshot artifacts into this repository. Default value is 'true'.`,
		},
		"suppress_pom_consistency_checks": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     suppressPom,
			Description: `(Optional) By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to 'true'. Default value is 'false'.`,
		},
		"reject_invalid_jars": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `(Optional) Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal". Default value is 'false'.`,
		},
	}, repoLayoutRefSchema("remote", repoType))

	type JavaRemoteRepo struct {
		RemoteRepositoryBaseParams
		FetchJarsEagerly             bool   `json:"fetchJarsEagerly"`
		FetchSourcesEagerly          bool   `json:"fetchSourcesEagerly"`
		RemoteRepoChecksumPolicyType string `json:"remoteRepoChecksumPolicyType"`
		HandleReleases               bool   `json:"handleReleases"`
		HandleSnapshots              bool   `json:"handleSnapshots"`
		SuppressPomConsistencyChecks bool   `json:"suppressPomConsistencyChecks"`
		RejectInvalidJars            bool   `json:"rejectInvalidJars"`
	}

	var unpackJavaRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{data}
		repo := JavaRemoteRepo{
			RemoteRepositoryBaseParams:   unpackBaseRemoteRepo(data, repoType),
			FetchJarsEagerly:             d.getBool("fetch_jars_eagerly", false),
			FetchSourcesEagerly:          d.getBool("fetch_sources_eagerly", false),
			RemoteRepoChecksumPolicyType: d.getString("remote_repo_checksum_policy_type", false),
			HandleReleases:               d.getBool("handle_releases", false),
			HandleSnapshots:              d.getBool("handle_snapshots", false),
			SuppressPomConsistencyChecks: d.getBool("suppress_pom_consistency_checks", false),
			RejectInvalidJars:            d.getBool("reject_invalid_jars", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(javaRemoteSchema, defaultPacker(javaRemoteSchema), unpackJavaRemoteRepo, func() interface{} {
		return &JavaRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: repoType,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}
	})
}
