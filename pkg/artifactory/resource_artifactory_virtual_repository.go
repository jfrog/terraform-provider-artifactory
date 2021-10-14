package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

var legacySchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{

	"debian_trivial_layout": {
		Type:     schema.TypeBool,
		Optional: true,
	},

	"key_pair": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"pom_repository_references_cleanup_policy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},

	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
})


var readFunc = mkRepoRead(packVirtualRepository, func() interface{} {
	return &MessyVirtualRepo{}
})

func resourceArtifactoryVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackVirtualRepository, readFunc),
		Read:   readFunc,
		Update: mkRepoUpdate(unpackVirtualRepository, readFunc),
		Delete: deleteRepo,
		Exists: repoExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: legacySchema,
		DeprecationMessage: "This resource is deprecated and you should use repo type specific resources " +
			"(such as artifactory_virtual_maven_repository) in the future",
	}
}

type MessyVirtualRepo struct {
	services.VirtualRepositoryBaseParams
	services.DebianVirtualRepositoryParams
	services.MavenVirtualRepositoryParams
	services.NugetVirtualRepositoryParams
}



func unpackVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := MessyVirtualRepo{
		VirtualRepositoryBaseParams: unpackBaseVirtRepo(s),
	}
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.KeyPair = d.getString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)

	return &repo, repo.Key, nil
}

func packVirtualRepository(r interface{}, d *schema.ResourceData) error {
	repo := r.(*MessyVirtualRepo)
	setValue := packBaseVirtRepo(d, repo.VirtualRepositoryBaseParams)

	setValue("debian_trivial_layout", repo.DebianTrivialLayout)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("key_pair", repo.KeyPair)
	setValue("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	setValue("repositories", repo.Repositories)
	errors := setValue("force_nuget_authentication", repo.ForceNugetAuthentication)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack virtual repo %q", errors)
	}

	return nil
}

