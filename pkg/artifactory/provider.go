package artifactory

import (
	"fmt"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/local"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/remote"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/security"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/virtual"
	"github.com/jfrog/terraform-provider-artifactory/pkg/xray"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Version for some reason isn't getting updated by the linker
var Version = "2.6.18"

// Provider Artifactory provider that supports configuration via username+password or a token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("ARTIFACTORY_URL", "http://localhost:8082"),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_USERNAME", nil),
				ValidateFunc:  validation.StringIsNotEmpty,
				ConflictsWith: []string{"api_key"},
				Deprecated:    "Xray and projects functionality will not work with any auth method other than access tokens (Bearer)",
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_PASSWORD", nil),
				ConflictsWith: []string{"access_token", "api_key"},
				ValidateFunc:  validation.StringIsNotEmpty,
				Deprecated:    "Xray and projects functionality will not work with any auth method other than access tokens (Bearer)",
				Description:   "Insider note: You may actually use an api_key as the password. This will get your around xray limitations instead of a bearer token",
			},
			"api_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_API_KEY", nil),
				ConflictsWith: []string{"username", "access_token", "password"},
				ValidateFunc:  validation.StringIsNotEmpty,
				Deprecated:    "Xray and projects functionality will not work with any auth method other than access tokens (Bearer)",
			},
			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_ACCESS_TOKEN", nil),
				ConflictsWith: []string{"api_key", "password"},
				Description:   "This is a bearer token that can be given to you by your admin under `Identity and Access`",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"artifactory_keypair":                    security.ResourceArtifactoryKeyPair(),
			"artifactory_local_repository":           local.ResourceArtifactoryLocalRepository(),
			"artifactory_local_nuget_repository":     local.ResourceArtifactoryLocalNugetRepository(),
			"artifactory_local_generic_repository":   local.ResourceArtifactoryLocalGenericRepository("generic"),
			"artifactory_local_npm_repository":       local.ResourceArtifactoryLocalGenericRepository("npm"),
			"artifactory_local_ivy_repository":       local.ResourceArtifactoryLocalGenericRepository("ivy"),
			"artifactory_local_sbt_repository":       local.ResourceArtifactoryLocalGenericRepository("sbt"),
			"artifactory_local_helm_repository":      local.ResourceArtifactoryLocalGenericRepository("helm"),
			"artifactory_local_cocoapods_repository": local.ResourceArtifactoryLocalGenericRepository("cocoapods"),
			"artifactory_local_opkg_repository":      local.ResourceArtifactoryLocalGenericRepository("opkg"),
			"artifactory_local_cran_repository":      local.ResourceArtifactoryLocalGenericRepository("cran"),
			"artifactory_local_gems_repository":      local.ResourceArtifactoryLocalGenericRepository("gems"),
			"artifactory_local_bower_repository":     local.ResourceArtifactoryLocalGenericRepository("bower"),
			"artifactory_local_composer_repository":  local.ResourceArtifactoryLocalGenericRepository("composer"),
			"artifactory_local_pypi_repository":      local.ResourceArtifactoryLocalGenericRepository("pypi"),
			"artifactory_local_vagrant_repository":   local.ResourceArtifactoryLocalGenericRepository("vagrant"),
			"artifactory_local_gitlfs_repository":    local.ResourceArtifactoryLocalGenericRepository("gitlfs"),
			"artifactory_local_go_repository":        local.ResourceArtifactoryLocalGenericRepository("go"),
			"artifactory_local_conan_repository":     local.ResourceArtifactoryLocalGenericRepository("conan"),
			"artifactory_local_chef_repository":      local.ResourceArtifactoryLocalGenericRepository("chef"),
			"artifactory_local_puppet_repository":    local.ResourceArtifactoryLocalGenericRepository("puppet"),
			"artifactory_local_maven_repository":     local.ResourceArtifactoryLocalJavaRepository("maven", false),
			"artifactory_local_gradle_repository":    local.ResourceArtifactoryLocalJavaRepository("gradle", true),
			"artifactory_local_alpine_repository":    local.ResourceArtifactoryLocalAlpineRepository(),
			"artifactory_local_debian_repository":    local.ResourceArtifactoryLocalDebianRepository(),
			"artifactory_local_docker_v2_repository": local.ResourceArtifactoryLocalDockerV2Repository(),
			"artifactory_local_docker_v1_repository": local.ResourceArtifactoryLocalDockerV1Repository(),
			"artifactory_local_rpm_repository":       local.ResourceArtifactoryLocalRpmRepository(),
			"artifactory_remote_repository":          remote.ResourceArtifactoryRemoteRepository(),
			"artifactory_remote_docker_repository":   remote.ResourceArtifactoryRemoteDockerRepository(),
			"artifactory_remote_helm_repository":     remote.ResourceArtifactoryRemoteHelmRepository(),
			"artifactory_remote_cargo_repository":    remote.ResourceArtifactoryRemoteCargoRepository(),
			"artifactory_virtual_repository":         virtual.ResourceArtifactoryVirtualRepository(),
			"artifactory_virtual_maven_repository":   virtual.ResourceArtifactoryMavenVirtualRepository(),
			"artifactory_virtual_go_repository":      virtual.ResourceArtifactoryGoVirtualRepository(),
			"artifactory_group":                      security.ResourceArtifactoryGroup(),
			"artifactory_user":                       security.ResourceArtifactoryUser(),
			"artifactory_permission_target":          security.ResourceArtifactoryPermissionTarget(),
			"artifactory_replication_config":         resourceArtifactoryReplicationConfig(),
			"artifactory_single_replication_config":  resourceArtifactorySingleReplicationConfig(),
			"artifactory_certificate":                security.ResourceArtifactoryCertificate(),
			"artifactory_api_key":                    security.ResourceArtifactoryApiKey(),
			"artifactory_access_token":               security.ResourceArtifactoryAccessToken(),
			"artifactory_general_security":           security.ResourceArtifactoryGeneralSecurity(),
			"artifactory_oauth_settings":             security.ResourceArtifactoryOauthSettings(),
			"artifactory_saml_settings":              security.ResourceArtifactorySamlSettings(),
			// Deprecated. Remove in V3
			"artifactory_permission_targets": security.ResourceArtifactoryPermissionTargets(),
			// Xray resources
			"artifactory_xray_policy": xray.ResourceXrayPolicy(),
			"artifactory_xray_watch":  xray.ResourceXrayWatch(),
		},

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

func addAuthToResty(client *resty.Client, username, password, apiKey, accessToken string) (*resty.Client, error) {
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
	}
	if apiKey != "" {
		return client.SetHeader("X-JFrog-Art-Api", apiKey), nil
	}
	if username != "" && password != "" {
		return client.SetBasicAuth(username, password), nil
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
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	restyBase, err = addAuthToResty(restyBase, username, password, apiKey, accessToken)
	if err != nil {
		return nil, err
	}
	_, err = sendUsageRepo(restyBase, terraformVersion)

	if err != nil {
		return nil, err
	}

	return restyBase, nil

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
