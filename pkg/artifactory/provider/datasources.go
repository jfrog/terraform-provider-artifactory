package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource"
	datasource_federated "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository/federated"
	datasource_local "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository/local"
	datasource_remote "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository/remote"
	datasource_virtual "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository/virtual"
	datasource_security "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/security"
	datasource_user "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/user"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/virtual"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func datasourcesMap() map[string]*schema.Resource {
	dataSourcesMap := map[string]*schema.Resource{
		"artifactory_file":                                    datasource.ArtifactoryFile(),
		"artifactory_fileinfo":                                datasource.ArtifactoryFileInfo(),
		"artifactory_group":                                   datasource_security.DataSourceArtifactoryGroup(),
		"artifactory_permission_target":                       datasource_security.DataSourceArtifactoryPermissionTarget(),
		"artifactory_user":                                    datasource_user.DataSourceArtifactoryUser(),
		"artifactory_local_alpine_repository":                 datasource_local.DataSourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":                  datasource_local.DataSourceArtifactoryLocalCargoRepository(),
		"artifactory_local_conan_repository":                  datasource_local.DataSourceArtifactoryLocalConanRepository(),
		"artifactory_local_debian_repository":                 datasource_local.DataSourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":              datasource_local.DataSourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":              datasource_local.DataSourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_maven_repository":                  datasource_local.DataSourceArtifactoryLocalJavaRepository("maven", false),
		"artifactory_local_nuget_repository":                  datasource_local.DataSourceArtifactoryLocalNugetRepository(),
		"artifactory_local_rpm_repository":                    datasource_local.DataSourceArtifactoryLocalRpmRepository(),
		"artifactory_local_terraform_module_repository":       datasource_local.DataSourceArtifactoryLocalTerraformRepository("module"),
		"artifactory_local_terraform_provider_repository":     datasource_local.DataSourceArtifactoryLocalTerraformRepository("provider"),
		"artifactory_remote_bower_repository":                 datasource_remote.DataSourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":                 datasource_remote.DataSourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository":             datasource_remote.DataSourceArtifactoryRemotecoCoapodsRepository(),
		"artifactory_remote_composer_repository":              datasource_remote.DataSourceArtifactoryRemotecoComposerRepository(),
		"artifactory_remote_conan_repository":                 datasource_remote.DataSourceArtifactoryRemotecoConanRepository(),
		"artifactory_remote_docker_repository":                datasource_remote.DataSourceArtifactoryRemotecoDockerRepository(),
		"artifactory_remote_generic_repository":               datasource_remote.DataSourceArtifactoryRemoteGenericRepository(),
		"artifactory_remote_go_repository":                    datasource_remote.DataSourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_helm_repository":                  datasource_remote.DataSourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_maven_repository":                 datasource_remote.DataSourceArtifactoryRemoteMavenRepository(),
		"artifactory_remote_npm_repository":                   datasource_remote.DataSourceArtifactoryRemoteNpmRepository(),
		"artifactory_remote_nuget_repository":                 datasource_remote.DataSourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_pypi_repository":                  datasource_remote.DataSourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_terraform_repository":             datasource_remote.DataSourceArtifactoryRemoteTerraformRepository(),
		"artifactory_remote_vcs_repository":                   datasource_remote.DataSourceArtifactoryRemoteVcsRepository(),
		"artifactory_virtual_alpine_repository":               datasource_virtual.DatasourceArtifactoryVirtualAlpineRepository(),
		"artifactory_virtual_bower_repository":                datasource_virtual.DatasourceArtifactoryVirtualBowerRepository(),
		"artifactory_virtual_debian_repository":               datasource_virtual.DatasourceArtifactoryVirtualDebianRepository(),
		"artifactory_virtual_conan_repository":                datasource_virtual.DatasourceArtifactoryVirtualConanRepository(),
		"artifactory_virtual_go_repository":                   datasource_virtual.DatasourceArtifactoryVirtualGoRepository(),
		"artifactory_virtual_docker_repository":               datasource_virtual.DatasourceArtifactoryVirtualDockerRepository(),
		"artifactory_virtual_helm_repository":                 datasource_virtual.DatasourceArtifactoryVirtualHelmRepository(),
		"artifactory_virtual_npm_repository":                  datasource_virtual.DatasourceArtifactoryVirtualNpmRepository(),
		"artifactory_virtual_nuget_repository":                datasource_virtual.DatasourceArtifactoryVirtualNugetRepository(),
		"artifactory_virtual_rpm_repository":                  datasource_virtual.DatasourceArtifactoryVirtualRpmRepository(),
		"artifactory_federated_alpine_repository":             datasource_federated.DataSourceArtifactoryFederatedAlpineRepository(),
		"artifactory_federated_cargo_repository":              datasource_federated.DataSourceArtifactoryFederatedCargoRepository(),
		"artifactory_federated_conan_repository":              datasource_federated.DataSourceArtifactoryFederatedConanRepository(),
		"artifactory_federated_debian_repository":             datasource_federated.DataSourceArtifactoryFederatedDebianRepository(),
		"artifactory_federated_docker_v1_repository":          datasource_federated.DataSourceArtifactoryFederatedDockerV1Repository(),
		"artifactory_federated_docker_v2_repository":          datasource_federated.DataSourceArtifactoryFederatedDockerV2Repository(),
		"artifactory_federated_docker_repository":             datasource_federated.DataSourceArtifactoryFederatedDockerV2Repository(),
		"artifactory_federated_maven_repository":              datasource_federated.DataSourceArtifactoryFederatedJavaRepository("maven", false),
		"artifactory_federated_nuget_repository":              datasource_federated.DataSourceArtifactoryFederatedNugetRepository(),
		"artifactory_federated_rpm_repository":                datasource_federated.DataSourceArtifactoryFederatedRpmRepository(),
		"artifactory_federated_terraform_module_repository":   datasource_federated.DataSourceArtifactoryFederatedTerraformRepository("module"),
		"artifactory_federated_terraform_provider_repository": datasource_federated.DataSourceArtifactoryFederatedTerraformRepository("provider"),
	}

	for _, packageType := range repository.GradleLikePackageTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
		dataSourcesMap[localResourceName] = datasource_local.DataSourceArtifactoryLocalJavaRepository(packageType, true)

		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
		dataSourcesMap[remoteResourceName] = datasource_remote.DataSourceArtifactoryRemoteJavaRepository(packageType, true)

		virtualDataSourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		dataSourcesMap[virtualDataSourceName] = datasource_virtual.DataSourceArtifactoryVirtualJavaRepository(packageType)

		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", packageType)
		dataSourcesMap[federatedResourceName] = datasource_federated.DataSourceArtifactoryFederatedJavaRepository(packageType, true)
	}

	for _, packageType := range local.PackageTypesLikeGeneric {
		localDataSourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
		dataSourcesMap[localDataSourceName] = datasource_local.DataSourceArtifactoryLocalGenericRepository(packageType)
	}

	for _, packageType := range remote.PackageTypesLikeBasic {
		remoteDataSourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
		dataSourcesMap[remoteDataSourceName] = datasource_remote.DataSourceArtifactoryRemoteBasicRepository(packageType)
	}

	for _, packageType := range virtual.PackageTypesLikeGeneric {
		virtualDataSourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		dataSourcesMap[virtualDataSourceName] = datasource_virtual.DataSourceArtifactoryVirtualGenericRepository(packageType)
	}
	for _, packageType := range virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualDataSourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		dataSourcesMap[virtualDataSourceName] = datasource_virtual.DataSourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType)
	}

	for _, packageType := range federated.PackageTypesLikeGeneric {
		federatedDataSourceName := fmt.Sprintf("artifactory_federated_%s_repository", packageType)
		dataSourcesMap[federatedDataSourceName] = datasource_federated.DataSourceArtifactoryFederatedGenericRepository(packageType)
	}

	return utilsdk.AddTelemetry(productId, dataSourcesMap)
}
