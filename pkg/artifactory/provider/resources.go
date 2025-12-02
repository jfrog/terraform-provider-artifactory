// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/replication"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
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

	for _, packageType := range repository.PackageTypesLikeGradle {
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
