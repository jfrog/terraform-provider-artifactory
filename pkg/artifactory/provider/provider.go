package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/replication"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

var Version = "6.6.0" // needs to be exported so make file can update this
var productId = "terraform-provider-artifactory/" + Version

// Provider Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() *schema.Provider {
	resourceMap := map[string]*schema.Resource{
		"artifactory_keypair":                     security.ResourceArtifactoryKeyPair(),
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
		"artifactory_group":                       security.ResourceArtifactoryGroup(),
		"artifactory_user":                        user.ResourceArtifactoryUser(),
		"artifactory_unmanaged_user":              user.ResourceArtifactoryUser(), // alias of artifactory_user
		"artifactory_managed_user":                user.ResourceArtifactoryManagedUser(),
		"artifactory_anonymous_user":              user.ResourceArtifactoryAnonymousUser(),
		"artifactory_permission_target":           security.ResourceArtifactoryPermissionTarget(),
		"artifactory_pull_replication":            replication.ResourceArtifactoryPullReplication(),
		"artifactory_push_replication":            replication.ResourceArtifactoryPushReplication(),
		"artifactory_certificate":                 security.ResourceArtifactoryCertificate(),
		"artifactory_api_key":                     security.ResourceArtifactoryApiKey(),
		"artifactory_access_token":                security.ResourceArtifactoryAccessToken(),
		"artifactory_general_security":            configuration.ResourceArtifactoryGeneralSecurity(),
		"artifactory_oauth_settings":              configuration.ResourceArtifactoryOauthSettings(),
		"artifactory_saml_settings":               configuration.ResourceArtifactorySamlSettings(),
		"artifactory_permission_targets":          security.ResourceArtifactoryPermissionTargets(), // Deprecated. Remove in V7
		"artifactory_replication_config":          replication.ResourceArtifactoryReplicationConfig(),
		"artifactory_single_replication_config":   replication.ResourceArtifactorySingleReplicationConfig(),
		"artifactory_ldap_setting":                configuration.ResourceArtifactoryLdapSetting(),
		"artifactory_ldap_group_setting":          configuration.ResourceArtifactoryLdapGroupSetting(),
		"artifactory_backup":                      configuration.ResourceArtifactoryBackup(),
	}

	for _, repoType := range local.RepoTypesLikeGeneric {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resourceMap[localResourceName] = local.ResourceArtifactoryLocalGenericRepository(repoType)
	}

	for _, repoType := range remote.RemoteRepoTypesLikeGeneric {
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resourceMap[remoteResourceName] = remote.ResourceArtifactoryRemoteGenericRepository(repoType)
	}

	for _, repoType := range repository.GradleLikeRepoTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resourceMap[localResourceName] = local.ResourceArtifactoryLocalJavaRepository(repoType, true)
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resourceMap[remoteResourceName] = remote.ResourceArtifactoryRemoteJavaRepository(repoType, true)
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourceMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualJavaRepository(repoType)
	}

	for _, repoType := range virtual.VirtualRepoTypesLikeGeneric {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourceMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualGenericRepository(repoType)
	}
	for _, repoType := range virtual.VirtualRepoTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resourceMap[virtualResourceName] = virtual.ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(repoType)
	}

	for _, repoType := range federated.FederatedRepoTypesSupported {
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
		resourceMap[federatedResourceName] = federated.ResourceArtifactoryFederatedGenericRepository(repoType)
	}

	for _, webhookType := range webhook.WebhookTypesSupported {
		webhookResourceName := fmt.Sprintf("artifactory_%s_webhook", webhookType)
		resourceMap[webhookResourceName] = webhook.ResourceArtifactoryWebhook(webhookType)
	}

	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ARTIFACTORY_URL", "JFROG_URL"}, "http://localhost:8082"),
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
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARTIFACTORY_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN"}, nil),
				Description: "This is a access token that can be given to you by your admin under `Identity and Access`. If not set, the 'api_key' attribute value will be used.",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
			},
		},

		ResourcesMap: 	addTelemetry(resourceMap),

		DataSourcesMap: addTelemetry(map[string]*schema.Resource{
			"artifactory_file":     datasource.ArtifactoryFile(),
			"artifactory_fileinfo": datasource.ArtifactoryFileInfo(),
		}),
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		tflog.Debug(ctx, "ConfigureContextFunc")
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}

	return p
}

func addTelemetry(resourceMap map[string]*schema.Resource) map[string]*schema.Resource {
	for name, skeema := range resourceMap {
		if skeema.Create != nil {
			panic(fmt.Sprintf("[%s] deprecated Create function in use", name))
		}
		if skeema.Read != nil {
			panic(fmt.Sprintf("[%s] deprecated Read function in use", name))
		}
		if skeema.Update != nil {
			panic(fmt.Sprintf("[%s] deprecated Update function in use", name))
		}
		if skeema.Delete != nil {
			panic(fmt.Sprintf("[%s] deprecated Delete function in use", name))
		}
	}

	for name, skeema := range resourceMap {
		if skeema.CreateContext != nil {
			skeema.CreateContext = util.ApplyTelemetry(Version, name, "CREATE", skeema.CreateContext)
		}
		if skeema.ReadContext != nil {
			skeema.ReadContext = util.ApplyTelemetry(Version, name, "READ", skeema.ReadContext)
		}
		if skeema.UpdateContext != nil {
			skeema.UpdateContext = util.ApplyTelemetry(Version, name, "UPDATE", skeema.UpdateContext)
		}
		if skeema.DeleteContext != nil {
			skeema.DeleteContext = util.ApplyTelemetry(Version, name, "DELETE", skeema.DeleteContext)
		}
	}
	return resourceMap
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	tflog.Debug(ctx, "providerConfigure")

	URL, ok := d.GetOk("url")
	if URL == nil || URL == "" || !ok {
		return nil, diag.Errorf("you must supply a URL")
	}

	restyBase, err := client.Build(URL.(string), Version)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	restyBase, err = client.AddAuth(restyBase, apiKey, accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	checkLicense := d.Get("check_license").(bool)
	if checkLicense {
		err = checkArtifactoryLicense(restyBase)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	util.SendUsage(ctx, restyBase, productId, featureUsage)

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
		return fmt.Errorf("failed to check for license. If your usage doesn't require admin permission, you can set `check_license` attribute to `false` to skip this check. %s", err)
	}

	var licenseType string
	if len(licensesWrapper.Licenses) > 0 {
		licenseType = licensesWrapper.Licenses[0].Type
	} else {
		licenseType = licensesWrapper.Type
	}

	if matched, _ := regexp.MatchString(`Enterprise|Commercial|Edge`, licenseType); !matched {
		return fmt.Errorf("artifactory requires Pro or Enterprise or Edge license to work with Terraform! If your usage doesn't require a license, you can set `check_license` attribute to `false` to skip this check")
	}

	return nil
}
