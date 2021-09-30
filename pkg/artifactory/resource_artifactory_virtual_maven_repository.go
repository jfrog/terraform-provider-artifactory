package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	},
	"key_pair": {
		Type:     schema.TypeString,
		Optional: true,
	},
})

func newMavenStruct() interface{} {
	return &services.MavenVirtualRepositoryParams{}
}

var mvnVirtReader = mkVirtualRepoRead(packMavenVirtualRepository, newMavenStruct)

func resourceArtifactoryMavenVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkVirtualCreate(unpackMavenVirtualRepository, mvnVirtReader),
		Read:   mvnVirtReader,
		Update: mkVirtualUpdate(unpackMavenVirtualRepository, mvnVirtReader),
		Delete: resourceVirtualRepositoryDelete,
		Exists: resourceVirtualRepositoryExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: mavenVirtualSchema,
	}
}

func unpackMavenVirtualRepository(s *schema.ResourceData) (interface{}, string) {
	d := &ResourceData{s}
	base, _ := unpackBaseVirtRepo(s)

	repo := services.MavenVirtualRepositoryParams{
		VirtualRepositoryBaseParams: base,
	}
	repo.KeyPair = d.getString("key_pair", false)
	repo.ForceMavenAuthentication = d.getBoolRef("force_maven_authentication", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)

	return repo, repo.Key
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
