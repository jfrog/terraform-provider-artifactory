package artifactory

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Version for some reason isn't getting updated by the linker
var Version = "2.6.18"

// Provider Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() *schema.Provider {
	resoucesMap := map[string]*schema.Resource{
		"artifactory_keypair":                     resourceArtifactoryKeyPair(),
		"artifactory_local_nuget_repository":      resourceArtifactoryLocalNugetRepository(),
		"artifactory_local_maven_repository":      resourceArtifactoryLocalJavaRepository("maven", false),
		"artifactory_local_alpine_repository":     resourceArtifactoryLocalAlpineRepository(),
		"artifactory_local_cargo_repository":      resourceArtifactoryLocalCargoRepository(),
		"artifactory_local_debian_repository":     resourceArtifactoryLocalDebianRepository(),
		"artifactory_local_docker_v2_repository":  resourceArtifactoryLocalDockerV2Repository(),
		"artifactory_local_docker_v1_repository":  resourceArtifactoryLocalDockerV1Repository(),
		"artifactory_local_rpm_repository":        resourceArtifactoryLocalRpmRepository(),
		"artifactory_remote_bower_repository":     resourceArtifactoryRemoteBowerRepository(),
		"artifactory_remote_cargo_repository":     resourceArtifactoryRemoteCargoRepository(),
		"artifactory_remote_cocoapods_repository": resourceArtifactoryRemoteCocoapodsRepository(),
		"artifactory_remote_composer_repository":  resourceArtifactoryRemoteComposerRepository(),
		"artifactory_remote_docker_repository":    resourceArtifactoryRemoteDockerRepository(),
		"artifactory_remote_go_repository":        resourceArtifactoryRemoteGoRepository(),
		"artifactory_remote_helm_repository":      resourceArtifactoryRemoteHelmRepository(),
		"artifactory_remote_maven_repository":     resourceArtifactoryRemoteJavaRepository("maven", false),
		"artifactory_remote_nuget_repository":     resourceArtifactoryRemoteNugetRepository(),
		"artifactory_remote_pypi_repository":      resourceArtifactoryRemotePypiRepository(),
		"artifactory_remote_vcs_repository":       resourceArtifactoryRemoteVcsRepository(),
		"artifactory_virtual_alpine_repository":   resourceArtifactoryAlpineVirtualRepository(),
		"artifactory_virtual_bower_repository":    resourceArtifactoryBowerVirtualRepository(),
		"artifactory_virtual_debian_repository":   resourceArtifactoryDebianVirtualRepository(),
		"artifactory_virtual_maven_repository":    resourceArtifactoryJavaVirtualRepository("maven"),
		"artifactory_virtual_nuget_repository":    resourceArtifactoryNugetVirtualRepository(),
		"artifactory_virtual_go_repository":       resourceArtifactoryGoVirtualRepository(),
		"artifactory_virtual_rpm_repository":      resourceArtifactoryRpmVirtualRepository(),
		"artifactory_virtual_helm_repository":     resourceArtifactoryHelmVirtualRepository(),
		"artifactory_group":                       resourceArtifactoryGroup(),
		"artifactory_user":                        resourceArtifactoryUser(),
		"artifactory_permission_target":           resourceArtifactoryPermissionTarget(),
		"artifactory_pull_replication":            resourceArtifactoryPullReplication(),
		"artifactory_push_replication":            resourceArtifactoryPushReplication(),
		"artifactory_certificate":                 resourceArtifactoryCertificate(),
		"artifactory_api_key":                     resourceArtifactoryApiKey(),
		"artifactory_access_token":                resourceArtifactoryAccessToken(),
		"artifactory_general_security":            resourceArtifactoryGeneralSecurity(),
		"artifactory_oauth_settings":              resourceArtifactoryOauthSettings(),
		"artifactory_saml_settings":               resourceArtifactorySamlSettings(),
		"artifactory_permission_targets":          resourceArtifactoryPermissionTargets(), // Deprecated. Remove in V7
		"artifactory_replication_config":          resourceArtifactoryReplicationConfig(),
		"artifactory_single_replication_config":   resourceArtifactorySingleReplicationConfig(),
		"artifactory_ldap_setting":                resourceArtifactoryLdapSetting(),
		"artifactory_ldap_group_setting":          resourceArtifactoryLdapGroupSetting(),
		"artifactory_backup":                      resourceArtifactoryBackup(),
	}

	for _, repoType := range repoTypesLikeGeneric {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resoucesMap[localResourceName] = resourceArtifactoryLocalGenericRepository(repoType)
	}

	for _, repoType := range remoteRepoTypesLikeGeneric {
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resoucesMap[remoteResourceName] = resourceArtifactoryRemoteGenericRepository(repoType)
	}

	for _, repoType := range gradleLikeRepoTypes {
		localResourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
		resoucesMap[localResourceName] = resourceArtifactoryLocalJavaRepository(repoType, true)
		remoteResourceName := fmt.Sprintf("artifactory_remote_%s_repository", repoType)
		resoucesMap[remoteResourceName] = resourceArtifactoryRemoteJavaRepository(repoType, true)
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = resourceArtifactoryJavaVirtualRepository(repoType)
	}

	for _, repoType := range virtualRepoTypesLikeGeneric {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = resourceArtifactoryVirtualGenericRepository(repoType)
	}
	for _, repoType := range virtualRepoTypesLikeGenericWithRetrievalCachePeriodSecs {
		virtualResourceName := fmt.Sprintf("artifactory_virtual_%s_repository", repoType)
		resoucesMap[virtualResourceName] = resourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(repoType)
	}

	for _, repoType := range federatedRepoTypesSupported {
		federatedResourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
		resoucesMap[federatedResourceName] = resourceArtifactoryFederatedGenericRepository(repoType)
	}

	for _, webhookType := range webhookTypesSupported {
		webhookResourceName := fmt.Sprintf("artifactory_%s_webhook", webhookType)
		resoucesMap[webhookResourceName] = resourceArtifactoryWebhook(webhookType)
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
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_ACCESS_TOKEN", nil),
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
			"artifactory_file":     dataSourceArtifactoryFile(),
			"artifactory_fileinfo": dataSourceArtifactoryFileInfo(),
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

func buildResty(URL string) (*resty.Client, error) {

	u, err := url.ParseRequestURI(URL)

	if err != nil {
		return nil, err
	}
	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	restyBase := resty.New().SetHostURL(baseUrl).OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		if response == nil {
			return fmt.Errorf("no response found")
		}

		if response.StatusCode() >= http.StatusBadRequest {
			return fmt.Errorf("\n%d %s %s\n%s", response.StatusCode(), response.Request.Method, response.Request.URL, string(response.Body()[:]))
		}
		return nil
	}).
		SetHeader("content-type", "application/json").
		SetHeader("accept", "*/*").
		SetHeader("user-agent", "jfrog/terraform-provider-artifactory:"+Version).
		SetRetryCount(5)
	restyBase.DisableWarn = true

	return restyBase, nil
}

func addAuthToResty(client *resty.Client, apiKey, accessToken string) (*resty.Client, error) {
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
	}
	if apiKey != "" {
		return client.SetHeader("X-JFrog-Art-Api", apiKey), nil
	}
	return nil, fmt.Errorf("no authentication details supplied")
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	URL, ok := d.GetOk("url")
	if URL == nil || URL == "" || !ok {
		return nil, fmt.Errorf("you must supply a URL")
	}

	restyBase, err := buildResty(URL.(string))
	if err != nil {
		return nil, err
	}
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	restyBase, err = addAuthToResty(restyBase, apiKey, accessToken)
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
