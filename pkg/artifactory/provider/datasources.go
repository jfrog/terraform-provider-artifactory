package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource"
	datasource_local "github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource/repository/local"
	datasource_remote "github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource/repository/remote"
	datasource_security "github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource/security"
	datasource_user "github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource/user"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/util"
)

func datasourcesMap() map[string]*schema.Resource {
	dataSourcesMap := map[string]*schema.Resource{
		"artifactory_file":                                datasource.ArtifactoryFile(),
		"artifactory_fileinfo":                            datasource.ArtifactoryFileInfo(),
		"artifactory_group":                               datasource_security.DataSourceArtifactoryGroup(),
		"artifactory_permission_target":                   datasource_security.DataSourceArtifactoryPermissionTarget(),
		"artifactory_user":                                datasource_user.DataSourceArtifactoryUser(),
		"artifactory_local_alpine_repository":             datasource_local.DataSourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":              datasource_local.DataSourceArtifactoryLocalCargoRepository(),
		"artifactory_local_debian_repository":             datasource_local.DataSourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":          datasource_local.DataSourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":          datasource_local.DataSourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_maven_repository":              datasource_local.DataSourceArtifactoryLocalJavaRepository("maven", false),
		"artifactory_local_nuget_repository":              datasource_local.DataSourceArtifactoryLocalNugetRepository(),
		"artifactory_local_rpm_repository":                datasource_local.DataSourceArtifactoryLocalRpmRepository(),
		"artifactory_local_terraform_module_repository":   datasource_local.DataSourceArtifactoryLocalTerraformRepository("module"),
		"artifactory_local_terraform_provider_repository": datasource_local.DataSourceArtifactoryLocalTerraformRepository("provider"),
		"artifactory_remote_bower_repository":             datasource_remote.DataSourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":             datasource_remote.DataSourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository":         datasource_remote.DataSourceArtifactoryRemotecoCoapodsRepository(),
		"artifactory_remote_composer_repository":          datasource_remote.DataSourceArtifactoryRemotecoComposerRepository(),
		"artifactory_remote_conan_repository":             datasource_remote.DataSourceArtifactoryRemotecoConanRepository(),
		"artifactory_remote_docker_repository":            datasource_remote.DataSourceArtifactoryRemotecoDockerRepository(),
		"artifactory_remote_generic_repository":           datasource_remote.DataSourceArtifactoryRemoteGenericRepository(),
		"artifactory_remote_go_repository":                datasource_remote.DataSourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_helm_repository":              datasource_remote.DataSourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_maven_repository":             datasource_remote.DataSourceArtifactoryRemoteMavenRepository(),
		"artifactory_remote_nuget_repository":             datasource_remote.DataSourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_pypi_repository":              datasource_remote.DataSourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_terraform_repository":         datasource_remote.DataSourceArtifactoryRemoteTerraformRepository(),
		"artifactory_remote_vcs_repository":               datasource_remote.DataSourceArtifactoryRemoteVcsRepository(),
	}

	for _, packageType := range repository.GradleLikePackageTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
		dataSourcesMap[localResourceName] = datasource_local.DataSourceArtifactoryLocalJavaRepository(packageType, true)

		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
		dataSourcesMap[remoteResourceName] = datasource_remote.DataSourceArtifactoryRemoteJavaRepository(packageType, true)
	}

	for _, packageType := range local.PackageTypesLikeGeneric {
		localDataSourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
		dataSourcesMap[localDataSourceName] = datasource_local.DataSourceArtifactoryLocalGenericRepository(packageType)
	}

	for _, packageType := range remote.PackageTypesLikeBasic {
		remoteDataSourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
		dataSourcesMap[remoteDataSourceName] = datasource_remote.DataSourceArtifactoryRemoteBasicRepository(packageType)
	}

	return util.AddTelemetry(productId, dataSourcesMap)
}
