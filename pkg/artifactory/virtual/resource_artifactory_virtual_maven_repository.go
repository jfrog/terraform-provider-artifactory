package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

var mavenVirtualSchema = util.MergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{

	"force_maven_authentication": {
		Type:        schema.TypeBool,
		Computed:    true,
		Optional:    true,
		Description: "User authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests.",
	},
	"pom_repository_references_cleanup_policy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.StringInSlice(
			[]string{"discard_active_reference", "discard_any_reference", "nothing"}, false,
		),
		Description: "(1: discard_active_reference) Discard Active References - Removes repository elements that are declared directly under project or under a profile in the same POM that is activeByDefault.\n" +
			"(2: discard_any_reference) Discard Any References - Removes all repository elements regardless of whether they are included in an active profile or not.\n" +
			"(3: nothing) Nothing - Does not remove any repository elements declared in the POM.",
	},
	"key_pair": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The keypair used to sign artifacts",
	},
})

type CommonMavenGradleVirtualRepositoryParams struct {
	ForceMavenAuthentication             bool   `json:"forceMavenAuthentication,omitempty"`
	PomRepositoryReferencesCleanupPolicy string `hcl:"pom_repository_references_cleanup_policy" json:"pomRepositoryReferencesCleanupPolicy,omitempty"`
	KeyPair                              string `hcl:"key_pair" json:"keyPair,omitempty"`
}

type MavenVirtualRepositoryParams struct {
	RepositoryBaseParams
	CommonMavenGradleVirtualRepositoryParams
}

func ResourceArtifactoryMavenVirtualRepository() *schema.Resource {
	return repos.MkResourceSchema(mavenVirtualSchema, util.DefaultPacker, unpackMavenVirtualRepository, func() interface{} {
		return &MavenVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: "maven",
			},
		}
	})

}

func unpackMavenVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: s}

	repo := MavenVirtualRepositoryParams{
		RepositoryBaseParams: unpackBaseVirtRepo(s),
		CommonMavenGradleVirtualRepositoryParams: CommonMavenGradleVirtualRepositoryParams{
			KeyPair:                              d.GetString("key_pair", false),
			ForceMavenAuthentication:             d.GetBool("force_maven_authentication", false),
			PomRepositoryReferencesCleanupPolicy: d.GetString("pom_repository_references_cleanup_policy", false),
		},
	}
	repo.PackageType = "maven"

	return &repo, repo.Key, nil
}
