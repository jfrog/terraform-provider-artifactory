package repository

import (
	"context"

	diagsdk "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schemasdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
)

var validRepositoryTypes = []string{"local", "remote", "virtual", "federated", "distribution"}
var validPackageTypes = []string{"alpine", "bower", "cargo", "chef", "cocoapods", "composer", "conan", "conda", "cran", "debian", "docker", "gems", "generic", "gitlfs", "go", "gradle", "helm", "huggingfaceml", "ivy", "maven", "npm", "nuget", "opkg", "p2", "pub", "puppet", "pypi", "rpm", "sbt", "swift", " terraform", "terraformbackend", "vagrant", "yum"}

func MkRepoReadDataSource(pack packer.PackFunc, construct repository.Constructor) schemasdk.ReadContextFunc {
	return func(ctx context.Context, d *schemasdk.ResourceData, m interface{}) diagsdk.Diagnostics {
		repo, err := construct()
		if err != nil {
			return diagsdk.FromErr(err)
		}

		key := d.Get("key").(string)
		// repo must be a pointer
		_, err = m.(utilsdk.ProvderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", key).
			Get(repository.RepositoriesEndpoint)

		if err != nil {
			return diagsdk.FromErr(err)
		}

		d.SetId(key)

		return diagsdk.FromErr(pack(repo, d))
	}
}
