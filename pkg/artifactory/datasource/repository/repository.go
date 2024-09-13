package repository

import (
	"context"
	"net/http"

	"github.com/jfrog/terraform-provider-shared/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
)

var validRepositoryTypes = []string{"local", "remote", "virtual", "federated", "distribution"}
var validPackageTypes = []string{"alpine", "bower", "cargo", "chef", "cocoapods", "composer", "conan", "conda", "cran", "debian", "docker", "gems", "generic", "gitlfs", "go", "gradle", "helm", "huggingfaceml", "ivy", "maven", "npm", "nuget", "opkg", "p2", "pub", "puppet", "pypi", "rpm", "sbt", "swift", " terraform", "terraformbackend", "vagrant", "yum"}

func MkRepoReadDataSource(pack packer.PackFunc, construct repository.Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, err := construct()
		if err != nil {
			return diag.FromErr(err)
		}

		key := d.Get("key").(string)
		// repo must be a pointer
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", key).
			Get(repository.RepositoriesEndpoint)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		if resp.IsError() {
			return diag.Errorf("%s", resp.String())
		}

		d.SetId(key)

		return diag.FromErr(pack(repo, d))
	}
}
