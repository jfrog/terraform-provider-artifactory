package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const AnsiblePackageType = "ansible"

type AnsibleRepo struct {
	RepositoryRemoteBaseParams
}

var AnsibleSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://galaxy.ansible.com",
				Description: "The remote repo URL. Default to 'https://galaxy.ansible.com'",
			},
		},
		repository.RepoLayoutRefSchema(rclass, AnsiblePackageType),
	)
}

func ResourceArtifactoryRemoteAnsibleRepository() *schema.Resource {
	var unpackRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := AnsibleRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, AnsiblePackageType),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, AnsiblePackageType)()
		if err != nil {
			return nil, err
		}

		return &AnsibleRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   AnsiblePackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	ansibleSchema := AnsibleSchema(true)

	return mkResourceSchema(ansibleSchema, packer.Default(ansibleSchema), unpackRepo, constructor)
}
