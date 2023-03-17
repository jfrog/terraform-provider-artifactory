package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/replication"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/util"
)

func resourcesMap() map[string]*schema.Resource {
	resourcesMap := map[string]*schema.Resource{
		"artifactory_keypair":                                 security.ResourceArtifactoryKeyPair(),
		"artifactory_federated_alpine_repository":             federated.ResourceArtifactoryFederatedAlpineRepository(),
		"artifactory_federated_cargo_repository":              federated.ResourceArtifactoryFederatedCargoRepository(),
		"artifactory_federated_debian_repository":             federated.ResourceArtifactoryFederatedDebianRepository(),
		"artifactory_federated_docker_repository":             federated.ResourceArtifactoryFederatedDockerV2Repository(), // Alias for backward compatibility
		"artifactory_federated_docker_v1_repository":          federated.ResourceArtifactoryFederatedDockerV1Repository(),
		"artifactory_federated_docker_v2_repository":          federated.ResourceArtifactoryFederatedDockerV2Repository(),
		"artifactory_federated_maven_repository":              federated.ResourceArtifactoryFederatedJavaRepository("maven", false),
		"artifactory_federated_nuget_repository":              federated.ResourceArtifactoryFederatedNugetRepository(),
		"artifactory_federated_rpm_repository":                federated.ResourceArtifactoryFederatedRpmRepository(),
		"artifactory_federated_terraform_module_repository":   federated.ResourceArtifactoryFederatedTerraformRepository("module"),
		"artifactory_federated_terraform_provider_repository": federated.ResourceArtifactoryFederatedTerraformRepository("provider"),
		"artifactory_local_nuget_repository":                  local.ResourceArtifactoryLocalNugetRepository(),
		"artifactory_local_maven_repository":                  local.ResourceArtifactoryLocalJavaRepository("maven", false),
		"artifactory_local_alpine_repository":                 local.ResourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":                  local.ResourceArtifactoryLocalCargoRepository(),
		"artifactory_local_debian_repository":                 local.ResourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":              local.ResourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":              local.ResourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_rpm_repository":                    local.ResourceArtifactoryLocalRpmRepository(),
		"artifactory_local_terraform_module_repository":       local.ResourceArtifactoryLocalTerraformRepository("module"),
		"artifactory_local_terraform_provider_repository":     local.ResourceArtifactoryLocalTerraformRepository("provider"),
		"artifactory_remote_bower_repository":                 remote.ResourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":                 remote.ResourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository":             remote.ResourceArtifactoryRemoteCocoapodsRepository(),
		"artifactory_remote_composer_repository":              remote.ResourceArtifactoryRemoteComposerRepository(),
		"artifactory_remote_conan_repository":                 remote.ResourceArtifactoryRemoteConanRepository(),
		"artifactory_remote_docker_repository":                remote.ResourceArtifactoryRemoteDockerRepository(),
		"artifactory_remote_generic_repository":               remote.ResourceArtifactoryRemoteGenericRepository(),
		"artifactory_remote_go_repository":                    remote.ResourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_helm_repository":                  remote.ResourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_maven_repository":                 remote.ResourceArtifactoryRemoteMavenRepository(),
		"artifactory_remote_nuget_repository":                 remote.ResourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_pypi_repository":                  remote.ResourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_terraform_repository":             remote.ResourceArtifactoryRemoteTerraformRepository(),
		"artifactory_remote_vcs_repository":                   remote.ResourceArtifactoryRemoteVcsRepository(),
		"artifactory_virtual_alpine_repository":               virtual.ResourceArtifactoryVirtualAlpineRepository(),
		"artifactory_virtual_bower_repository":                virtual.ResourceArtifactoryVirtualBowerRepository(),
		"artifactory_virtual_debian_repository":               virtual.ResourceArtifactoryVirtualDebianRepository(),
		"artifactory_virtual_docker_repository":               virtual.ResourceArtifactoryVirtualDockerRepository(),
		"artifactory_virtual_maven_repository":                virtual.ResourceArtifactoryVirtualJavaRepository("maven"),
		"artifactory_virtual_npm_repository":                  virtual.ResourceArtifactoryVirtualNpmRepository(),
		"artifactory_virtual_nuget_repository":                virtual.ResourceArtifactoryVirtualNugetRepository(),
		"artifactory_virtual_go_repository":                   virtual.ResourceArtifactoryVirtualGoRepository(),
		"artifactory_virtual_rpm_repository":                  virtual.ResourceArtifactoryVirtualRpmRepository(),
		"artifactory_virtual_helm_repository":                 virtual.ResourceArtifactoryVirtualHelmRepository(),
		"artifactory_group":                                   security.ResourceArtifactoryGroup(),
		"artifactory_user":                                    user.ResourceArtifactoryUser(),
		"artifactory_unmanaged_user":                          user.ResourceArtifactoryUser(), // alias of artifactory_user
		"artifactory_managed_user":                            user.ResourceArtifactoryManagedUser(),
		"artifactory_anonymous_user":                          user.ResourceArtifactoryAnonymousUser(),
		"artifactory_permission_target":                       security.ResourceArtifactoryPermissionTarget(),
		"artifactory_pull_replication":                        replication.ResourceArtifactoryPullReplication(),
		"artifactory_push_replication":                        replication.ResourceArtifactoryPushReplication(),
		"artifactory_local_repository_single_replication":     replication.ResourceArtifactoryLocalRepositorySingleReplication(),
		"artifactory_local_repository_multi_replication":      replication.ResourceArtifactoryLocalRepositoryMultiReplication(),
		"artifactory_remote_repository_replication":           replication.ResourceArtifactoryRemoteRepositoryReplication(),
		"artifactory_certificate":                             security.ResourceArtifactoryCertificate(),
		"artifactory_api_key":                                 security.ResourceArtifactoryApiKey(),
		"artifactory_access_token":                            security.ResourceArtifactoryAccessToken(),
		"artifactory_scoped_token":                            security.ResourceArtifactoryScopedToken(),
		"artifactory_general_security":                        configuration.ResourceArtifactoryGeneralSecurity(),
		"artifactory_oauth_settings":                          configuration.ResourceArtifactoryOauthSettings(),
		"artifactory_saml_settings":                           configuration.ResourceArtifactorySamlSettings(),
		"artifactory_permission_targets":                      security.ResourceArtifactoryPermissionTargets(), // Deprecated. Remove in V7
		"artifactory_replication_config":                      replication.ResourceArtifactoryReplicationConfig(),
		"artifactory_single_replication_config":               replication.ResourceArtifactorySingleReplicationConfig(),
		"artifactory_ldap_setting":                            configuration.ResourceArtifactoryLdapSetting(),
		"artifactory_ldap_group_setting":                      configuration.ResourceArtifactoryLdapGroupSetting(),
		"artifactory_backup":                                  configuration.ResourceArtifactoryBackup(),
		"artifactory_repository_layout":                       configuration.ResourceArtifactoryRepositoryLayout(),
		"artifactory_property_set":                            configuration.ResourceArtifactoryPropertySet(),
		"artifactory_proxy":                                   configuration.ResourceArtifactoryProxy(),
	}

	for _, repoType := range local.PackageTypesLikeGeneric {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resourcesMap[localResourceName] = local.ResourceArtifactoryLocalGenericRepository(repoType)
	}

	for _, repoType := range remote.PackageTypesLikeBasic {
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resourcesMap[remoteResourceName] = remote.ResourceArtifactoryRemoteBasicRepository(repoType)
	}

	for _, repoType := range repository.GradleLikePackageTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resourcesMap[localResourceName] = local.ResourceArtifactoryLocalJavaRepository(repoType, true)
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resourcesMap[remoteResourceName] = remote.ResourceArtifactoryRemoteJavaRepository(repoType, true)
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualJavaRepository(repoType)
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
		resourcesMap[federatedResourceName] = federated.ResourceArtifactoryFederatedJavaRepository(repoType, true)
	}

	for _, repoType := range virtual.PackageTypesLikeGeneric {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualGenericRepository(repoType)
	}
	for _, repoType := range virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourcesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(repoType)
	}

	for _, repoType := range federated.PackageTypesLikeGeneric {
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
		resourcesMap[federatedResourceName] = federated.ResourceArtifactoryFederatedGenericRepository(repoType)
	}

	for _, webhookType := range webhook.TypesSupported {
		webhookResourceName := fmt.Sprintf("artifactory_%s_webhook", webhookType)
		resourcesMap[webhookResourceName] = webhook.ResourceArtifactoryWebhook(webhookType)
	}

	return util.AddTelemetry(productId, resourcesMap)
}
