package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type AnsibleRepo struct {
	RepositoryRemoteBaseParams
}

var ansibleSchema = lo.Assign(
	BaseSchema,
	map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://galaxy.ansible.com",
			Description: "The remote repo URL. Default to 'https://galaxy.ansible.com'",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AnsiblePackageType),
)

var AnsibleSchemas = GetSchemas(ansibleSchema)

func ResourceArtifactoryRemoteAnsibleRepository() *schema.Resource {
	var unpackRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := AnsibleRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.AnsiblePackageType),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.AnsiblePackageType)
		if err != nil {
			return nil, err
		}

		return &AnsibleRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.AnsiblePackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		AnsibleSchemas,
		packer.Default(AnsibleSchemas[CurrentSchemaVersion]),
		unpackRepo,
		constructor,
	)
}
