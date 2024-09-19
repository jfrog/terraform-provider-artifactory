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
var validPackageTypes = []string{
	repository.AlpinePackageType,
	repository.BowerPackageType,
	repository.CargoPackageType,
	repository.ChefPackageType,
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.ConanPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
	repository.DebianPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GitLFSPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HelmPackageType,
	repository.HuggingFacePackageType,
	repository.IvyPackageType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.OpkgPackageType,
	repository.P2PackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.RPMPackageType,
	repository.SBTPackageType,
	repository.SwiftPackageType,
	repository.TerraformPackageType,
	repository.TerraformBackendPackageType,
	repository.VagrantPackageType,
}

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
