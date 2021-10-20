package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

var mavenVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{

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
		Type:     schema.TypeString,
		Optional: true,
	},
})

var mvnVirtReader = mkRepoRead(packMavenVirtualRepository, func() interface{} {
	return &services.MavenVirtualRepositoryParams{
		VirtualRepositoryBaseParams: services.VirtualRepositoryBaseParams{
			Rclass:      "virtual",
			PackageType: "maven",
		}}
})

func resourceArtifactoryMavenVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackMavenVirtualRepository, mvnVirtReader),
		Read:   mvnVirtReader,
		Update: mkRepoUpdate(unpackMavenVirtualRepository, mvnVirtReader),
		Delete: deleteRepo,
		Exists: repoExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: mavenVirtualSchema,
	}
}

func unpackMavenVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}

	repo := services.MavenVirtualRepositoryParams{
		VirtualRepositoryBaseParams: unpackBaseVirtRepo(s),
		CommonMavenGradleVirtualRepositoryParams: services.CommonMavenGradleVirtualRepositoryParams{
			KeyPair:                              d.getString("key_pair", false),
			ForceMavenAuthentication:             d.getBoolRef("force_maven_authentication", false),
			PomRepositoryReferencesCleanupPolicy: d.getString("pom_repository_references_cleanup_policy", false),
		},
	}
	repo.PackageType = "maven"

	return &repo, repo.Key, nil
}

func packMavenVirtualRepository(r interface{}, d *schema.ResourceData) error {
	repo := r.(*services.MavenVirtualRepositoryParams)
	setValue := packBaseVirtRepo(d, repo.VirtualRepositoryBaseParams)

	setValue("key_pair", repo.KeyPair)
	setValue("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	errors := setValue("force_maven_authentication", *repo.ForceMavenAuthentication)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack virtual repo %q", errors)
	}

	return nil
}
