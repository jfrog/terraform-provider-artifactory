package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/replication"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func resourcesMap() map[string]*schema.Resource {
	resourcesMap := map[string]*schema.Resource{
		"artifactory_federated_alpine_repository":             federated.ResourceArtifactoryFederatedAlpineRepository(),
		"artifactory_federated_ansible_repository":            federated.ResourceArtifactoryFederatedAnsibleRepository(),
		"artifactory_federated_cargo_repository":              federated.ResourceArtifactoryFederatedCargoRepository(),
		"artifactory_federated_conan_repository":              federated.ResourceArtifactoryFederatedConanRepository(),
		"artifactory_federated_debian_repository":             federated.ResourceArtifactoryFederatedDebianRepository(),
		"artifactory_federated_docker_repository":             federated.ResourceArtifactoryFederatedDockerV2Repository(), // Alias for backward compatibility
		"artifactory_federated_docker_v1_repository":          federated.ResourceArtifactoryFederatedDockerV1Repository(),
		"artifactory_federated_docker_v2_repository":          federated.ResourceArtifactoryFederatedDockerV2Repository(),
		"artifactory_federated_helmoci_repository":            federated.ResourceArtifactoryFederatedHelmOciRepository(),
		"artifactory_federated_maven_repository":              federated.ResourceArtifactoryFederatedJavaRepository(repository.MavenPackageType, false),
		"artifactory_federated_nuget_repository":              federated.ResourceArtifactoryFederatedNugetRepository(),
		"artifactory_federated_oci_repository":                federated.ResourceArtifactoryFederatedOciRepository(),
		"artifactory_federated_rpm_repository":                federated.ResourceArtifactoryFederatedRpmRepository(),
		"artifactory_federated_terraform_module_repository":   federated.ResourceArtifactoryFederatedTerraformRepository("module"),
		"artifactory_federated_terraform_provider_repository": federated.ResourceArtifactoryFederatedTerraformRepository("provider"),
		"artifactory_local_ansible_repository":                local.ResourceArtifactoryLocalAnsibleRepository(),
		"artifactory_local_alpine_repository":                 local.ResourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":                  local.ResourceArtifactoryLocalCargoRepository(),
		"artifactory_local_conan_repository":                  local.ResourceArtifactoryLocalConanRepository(),
		"artifactory_local_debian_repository":                 local.ResourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":              local.ResourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":              local.ResourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_helmoci_repository":                local.ResourceArtifactoryLocalHelmOciRepository(),
		"artifactory_local_maven_repository":                  local.ResourceArtifactoryLocalJavaRepository(repository.MavenPackageType, false),
		"artifactory_local_nuget_repository":                  local.ResourceArtifactoryLocalNugetRepository(),
		"artifactory_local_oci_repository":                    local.ResourceArtifactoryLocalOciRepository(),
		"artifactory_local_rpm_repository":                    local.ResourceArtifactoryLocalRpmRepository(),
		"artifactory_local_terraform_module_repository":       local.ResourceArtifactoryLocalTerraformRepository("module"),
		"artifactory_local_terraform_provider_repository":     local.ResourceArtifactoryLocalTerraformRepository("provider"),
		"artifactory_remote_ansible_repository":               remote.ResourceArtifactoryRemoteAnsibleRepository(),
		"artifactory_remote_bower_repository":                 remote.ResourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":                 remote.ResourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository":             remote.ResourceArtifactoryRemoteCocoapodsRepository(),
		"artifactory_remote_composer_repository":              remote.ResourceArtifactoryRemoteComposerRepository(),
		"artifactory_remote_conan_repository":                 remote.ResourceArtifactoryRemoteConanRepository(),
		"artifactory_remote_docker_repository":                remote.ResourceArtifactoryRemoteDockerRepository(),
		"artifactory_remote_gems_repository":                  remote.ResourceArtifactoryRemoteGemsRepository(),
		"artifactory_remote_generic_repository":               remote.ResourceArtifactoryRemoteGenericRepository(),
		"artifactory_remote_go_repository":                    remote.ResourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_gradle_repository":                remote.ResourceArtifactoryRemoteGradleRepository(),
		"artifactory_remote_helm_repository":                  remote.ResourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_helmoci_repository":               remote.ResourceArtifactoryRemoteHelmOciRepository(),
		"artifactory_remote_huggingfaceml_repository":         remote.ResourceArtifactoryRemoteHuggingFaceRepository(),
		"artifactory_remote_ivy_repository":                   remote.ResourceArtifactoryRemoteJavaRepository(repository.IvyPackageType, true),
		"artifactory_remote_maven_repository":                 remote.ResourceArtifactoryRemoteMavenRepository(),
		"artifactory_remote_npm_repository":                   remote.ResourceArtifactoryRemoteNpmRepository(),
		"artifactory_remote_nuget_repository":                 remote.ResourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_oci_repository":                   remote.ResourceArtifactoryRemoteOciRepository(),
		"artifactory_remote_pypi_repository":                  remote.ResourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_sbt_repository":                   remote.ResourceArtifactoryRemoteJavaRepository(repository.SBTPackageType, true),
		"artifactory_remote_terraform_repository":             remote.ResourceArtifactoryRemoteTerraformRepository(),
		"artifactory_remote_vcs_repository":                   remote.ResourceArtifactoryRemoteVcsRepository(),
		"artifactory_virtual_alpine_repository":               virtual.ResourceArtifactoryVirtualAlpineRepository(),
		"artifactory_virtual_bower_repository":                virtual.ResourceArtifactoryVirtualBowerRepository(),
		"artifactory_virtual_conan_repository":                virtual.ResourceArtifactoryVirtualConanRepository(),
		"artifactory_virtual_debian_repository":               virtual.ResourceArtifactoryVirtualDebianRepository(),
		"artifactory_virtual_docker_repository":               virtual.ResourceArtifactoryVirtualDockerRepository(),
		"artifactory_virtual_go_repository":                   virtual.ResourceArtifactoryVirtualGoRepository(),
		"artifactory_virtual_helm_repository":                 virtual.ResourceArtifactoryVirtualHelmRepository(),
		"artifactory_virtual_helmoci_repository":              virtual.ResourceArtifactoryVirtualHelmOciRepository(),
		"artifactory_virtual_maven_repository":                virtual.ResourceArtifactoryVirtualJavaRepository(repository.MavenPackageType),
		"artifactory_virtual_npm_repository":                  virtual.ResourceArtifactoryVirtualNpmRepository(),
		"artifactory_virtual_nuget_repository":                virtual.ResourceArtifactoryVirtualNugetRepository(),
		"artifactory_virtual_oci_repository":                  virtual.ResourceArtifactoryVirtualOciRepository(),
		"artifactory_virtual_rpm_repository":                  virtual.ResourceArtifactoryVirtualRpmRepository(),
		"artifactory_permission_target":                       security.ResourceArtifactoryPermissionTarget(),
		"artifactory_pull_replication":                        replication.ResourceArtifactoryPullReplication(),
		"artifactory_push_replication":                        replication.ResourceArtifactoryPushReplication(),
		"artifactory_api_key":                                 security.ResourceArtifactoryApiKey(),
		"artifactory_oauth_settings":                          configuration.ResourceArtifactoryOauthSettings(),
		"artifactory_saml_settings":                           configuration.ResourceArtifactorySamlSettings(),
		"artifactory_ldap_setting":                            configuration.ResourceArtifactoryLdapSetting(),
		"artifactory_ldap_group_setting":                      configuration.ResourceArtifactoryLdapGroupSetting(),
	}

	for _, packageType := range remote.PackageTypesLikeBasic {
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
		resourcesMap[remoteResourceName] = remote.ResourceArtifactoryRemoteBasicRepository(packageType)
	}

	for _, packageType := range repository.PackageTypesLikeGradle {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
		resourcesMap[localResourceName] = local.ResourceArtifactoryLocalJavaRepository(packageType, true)
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualJavaRepository(packageType)
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", packageType)
		resourcesMap[federatedResourceName] = federated.ResourceArtifactoryFederatedJavaRepository(packageType, true)
	}

	for _, packageType := range virtual.PackageTypesLikeGeneric {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualGenericRepository(packageType)
	}

	for _, packageType := range virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", packageType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType)
	}

	for _, packageType := range federated.PackageTypesLikeGeneric {
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", packageType)
		resourcesMap[federatedResourceName] = federated.ResourceArtifactoryFederatedGenericRepository(packageType)
	}

	return utilsdk.AddTelemetry(productId, resourcesMap)
}
