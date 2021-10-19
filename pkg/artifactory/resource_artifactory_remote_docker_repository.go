package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dockerRemoteSchema = mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"external_dependencies_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Also known as 'Foreign Layers Caching' on the UI",
	},
	"enable_token_authentication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Enable token (Bearer) based authentication.",
	},
	"block_pushing_schema1": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "When set, Artifactory will block the pulling of Docker images with manifest v2 schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1 that exist in the cache.",
	},
	"external_dependencies_patterns": {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		RequiredWith: []string{"external_dependencies_enabled"},
		Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
			"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
			"By default, this is set to '**', which means that remote modules may be downloaded from any external VCS source.",
	},
})

type DockerRemoteRepo struct {
	RemoteRepositoryBaseParams
	ExternalDependenciesEnabled  *bool    `json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    *bool    `json:"enableTokenAuthentication,omitempty"`
	BlockPushingSchema1          *bool    `json:"blockPushingSchema1,omitempty"`
}

var dockerRemoteRepoReadFun = mkRepoRead(packDockerRemoteRepo, func() interface{} {
	return &DockerRemoteRepo{
		RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
			Rclass: "remote",
			PackageType: "docker",
		},
	}
})

func resourceArtifactoryRemoteDockerRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackDockerRemoteRepo, dockerRemoteRepoReadFun),
		Read:   dockerRemoteRepoReadFun,
		Update: mkRepoUpdate(unpackDockerRemoteRepo, dockerRemoteRepoReadFun),
		Delete: deleteRepo,
		Exists: repoExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: dockerRemoteSchema,
	}
}

func unpackDockerRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := DockerRemoteRepo{
		RemoteRepositoryBaseParams:   unpackBaseRemoteRepo(s),
		EnableTokenAuthentication:    d.getBoolRef("enable_token_authentication", false),
		ExternalDependenciesEnabled:  d.getBoolRef("external_dependencies_enabled", false),
		BlockPushingSchema1:          d.getBoolRef("block_pushing_schema1", false),
		ExternalDependenciesPatterns: d.getList("external_dependencies_patterns"),
	}

	return repo, repo.Key, nil
}

func packDockerRemoteRepo(r interface{}, d *schema.ResourceData) error {
	repo := r.(*DockerRemoteRepo)
	setValue := packBaseRemoteRepo(d, repo.RemoteRepositoryBaseParams)

	setValue("enable_token_authentication", *repo.EnableTokenAuthentication)
	setValue("external_dependencies_enabled", *repo.ExternalDependenciesEnabled)
	setValue("external_dependencies_patterns", repo.ExternalDependenciesPatterns)
	errors := setValue("block_pushing_schema1", *repo.BlockPushingSchema1)

	if len(errors) > 0 {
		return fmt.Errorf("%q", errors)
	}

	return nil
}
