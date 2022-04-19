package provider

import (
	"fmt"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

// Version for some reason isn't getting updated by the linker
var Version = "2.6.18"

// Provider Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() *schema.Provider {
	resoucesMap := map[string]*schema.Resource{
		"artifactory_keypair":                     artifactory.ResourceArtifactoryKeyPair(),
		"artifactory_local_nuget_repository":      local.ResourceArtifactoryLocalNugetRepository(),
		"artifactory_local_maven_repository":      local.ResourceArtifactoryLocalJavaRepository("maven", false),
		"artifactory_local_alpine_repository":     local.ResourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":      local.ResourceArtifactoryLocalCargoRepository(),
		"artifactory_local_debian_repository":     local.ResourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":  local.ResourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":  local.ResourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_rpm_repository":        local.ResourceArtifactoryLocalRpmRepository(),
		"artifactory_remote_bower_repository":     remote.ResourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":     remote.ResourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository": remote.ResourceArtifactoryRemoteCocoapodsRepository(),
		"artifactory_remote_composer_repository":  remote.ResourceArtifactoryRemoteComposerRepository(),
		"artifactory_remote_docker_repository":    remote.ResourceArtifactoryRemoteDockerRepository(),
		"artifactory_remote_go_repository":        remote.ResourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_helm_repository":      remote.ResourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_maven_repository":     remote.ResourceArtifactoryRemoteJavaRepository("maven", false),
		"artifactory_remote_nuget_repository":     remote.ResourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_pypi_repository":      remote.ResourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_vcs_repository":       remote.ResourceArtifactoryRemoteVcsRepository(),
		"artifactory_virtual_alpine_repository":   virtual.ResourceArtifactoryVirtualAlpineRepository(),
		"artifactory_virtual_bower_repository":    virtual.ResourceArtifactoryVirtualBowerRepository(),
		"artifactory_virtual_debian_repository":   virtual.ResourceArtifactoryVirtualDebianRepository(),
		"artifactory_virtual_maven_repository":    virtual.ResourceArtifactoryVirtualJavaRepository("maven"),
		"artifactory_virtual_nuget_repository":    virtual.ResourceArtifactoryVirtualNugetRepository(),
		"artifactory_virtual_go_repository":       virtual.ResourceArtifactoryVirtualGoRepository(),
		"artifactory_virtual_rpm_repository":      virtual.ResourceArtifactoryVirtualRpmRepository(),
		"artifactory_virtual_helm_repository":     virtual.ResourceArtifactoryVirtualHelmRepository(),
		"artifactory_group":                       artifactory.ResourceArtifactoryGroup(),
		"artifactory_user":                        artifactory.ResourceArtifactoryUser(),
		"artifactory_unmanaged_user":              artifactory.ResourceArtifactoryUser(), // alias of artifactory_user
		"artifactory_managed_user":                artifactory.ResourceArtifactoryManagedUser(),
		"artifactory_anonymous_user":              artifactory.ResourceArtifactoryAnonymousUser(),
		"artifactory_permission_target":           artifactory.ResourceArtifactoryPermissionTarget(),
		"artifactory_pull_replication":            artifactory.ResourceArtifactoryPullReplication(),
		"artifactory_push_replication":            artifactory.ResourceArtifactoryPushReplication(),
		"artifactory_certificate":                 artifactory.ResourceArtifactoryCertificate(),
		"artifactory_api_key":                     artifactory.ResourceArtifactoryApiKey(),
		"artifactory_access_token":                artifactory.ResourceArtifactoryAccessToken(),
		"artifactory_general_security":            artifactory.ResourceArtifactoryGeneralSecurity(),
		"artifactory_oauth_settings":              artifactory.ResourceArtifactoryOauthSettings(),
		"artifactory_saml_settings":               artifactory.ResourceArtifactorySamlSettings(),
		"artifactory_permission_targets":          artifactory.ResourceArtifactoryPermissionTargets(), // Deprecated. Remove in V7
		"artifactory_replication_config":          artifactory.ResourceArtifactoryReplicationConfig(),
		"artifactory_single_replication_config":   artifactory.ResourceArtifactorySingleReplicationConfig(),
		"artifactory_ldap_setting":                artifactory.ResourceArtifactoryLdapSetting(),
		"artifactory_ldap_group_setting":          artifactory.ResourceArtifactoryLdapGroupSetting(),
		"artifactory_backup":                      artifactory.ResourceArtifactoryBackup(),
	}

	for _, repoType := range local.RepoTypesLikeGeneric {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resoucesMap[localResourceName] = local.ResourceArtifactoryLocalGenericRepository(repoType)
	}

	for _, repoType := range remote.RemoteRepoTypesLikeGeneric {
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resoucesMap[remoteResourceName] = remote.ResourceArtifactoryRemoteGenericRepository(repoType)
	}

	for _, repoType := range utils.GradleLikeRepoTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resoucesMap[localResourceName] = local.ResourceArtifactoryLocalJavaRepository(repoType, true)
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resoucesMap[remoteResourceName] = remote.ResourceArtifactoryRemoteJavaRepository(repoType, true)
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualJavaRepository(repoType)
	}

	for _, repoType := range virtual.VirtualRepoTypesLikeGeneric {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualGenericRepository(repoType)
	}
	for _, repoType := range virtual.VirtualRepoTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(repoType)
	}

	for _, repoType := range federated.FederatedRepoTypesSupported {
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
		resoucesMap[federatedResourceName] = federated.ResourceArtifactoryFederatedGenericRepository(repoType)
	}

	for _, webhookType := range webhook.WebhookTypesSupported {
		webhookResourceName := fmt.Sprintf("artifactory_%s_webhook", webhookType)
		resoucesMap[webhookResourceName] = webhook.ResourceArtifactoryWebhook(webhookType)
	}

	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("ARTIFACTORY_URL", "http://localhost:8082"),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"api_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_API_KEY", nil),
				ConflictsWith: []string{"access_token"},
				ValidateFunc:  validation.StringIsNotEmpty,
				Description:   "API token. Projects functionality will not work with any auth method other than access tokens",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARTIFACTORY_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN"}, ""),
				Description: "This is a access token that can be given to you by your admin under `Identity and Access`",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
			},
		},

		ResourcesMap: resoucesMap,

		DataSourcesMap: map[string]*schema.Resource{
			"artifactory_file":     artifactory.DataSourceArtifactoryFile(),
			"artifactory_fileinfo": artifactory.DataSourceArtifactoryFileInfo(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return p
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	URL, ok := d.GetOk("url")
	if URL == nil || URL == "" || !ok {
		return nil, fmt.Errorf("you must supply a URL")
	}

	restyBase, err := utils.BuildResty(URL.(string), Version)
	if err != nil {
		return nil, err
	}
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	restyBase, err = utils.AddAuthToResty(restyBase, apiKey, accessToken)
	if err != nil {
		return nil, err
	}

	checkLicense := d.Get("check_license").(bool)
	if checkLicense {
		err = checkArtifactoryLicense(restyBase)
		if err != nil {
			return nil, err
		}
	}

	_, err = sendUsageRepo(restyBase, terraformVersion)

	if err != nil {
		return nil, err
	}

	return restyBase, nil
}

func checkArtifactoryLicense(client *resty.Client) error {

	type License struct {
		Type string `json:"type"`
	}

	type LicensesWrapper struct {
		License
		Licenses []License `json:"licenses"` // HA licenses returns as an array instead
	}

	licensesWrapper := LicensesWrapper{}
	_, err := client.R().
		SetResult(&licensesWrapper).
		Get("/artifactory/api/system/license")

	if err != nil {
		return fmt.Errorf("Failed to check for license. If your usage doesn't require admin permission, you can set `check_license` attribute to `false` to skip this check. %s", err)
	}

	var licenseType string
	if len(licensesWrapper.Licenses) > 0 {
		licenseType = licensesWrapper.Licenses[0].Type
	} else {
		licenseType = licensesWrapper.Type
	}

	if matched, _ := regexp.MatchString(`(?:Enterprise|Commercial|Edge)`, licenseType); !matched {
		return fmt.Errorf("Artifactory requires Pro or Enterprise or Edge license to work with Terraform! If your usage doesn't require a license, you can set `check_license` attribute to `false` to skip this check.")
	}

	return nil
}

func sendUsageRepo(restyBase *resty.Client, terraformVersion string) (interface{}, error) {
	type Feature struct {
		FeatureId string `json:"featureId"`
	}
	type UsageStruct struct {
		ProductId string    `json:"productId"`
		Features  []Feature `json:"features"`
	}
	_, err := restyBase.R().SetBody(UsageStruct{
		"terraform-provider-artifactory/" + Version,
		[]Feature{
			{FeatureId: "Partner/ACC-007450"},
			{FeatureId: "Terraform/" + terraformVersion},
		},
	}).Post("artifactory/api/system/usage")

	if err != nil {
		return nil, fmt.Errorf("unable to report usage %s", err)
	}
	return nil, nil
}
