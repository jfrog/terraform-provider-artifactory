package artifactory

import (
	"fmt"
	artifactoryold "github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/jasonwbarnett/go-xray/xray"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"net/http"
	"net/url"
)

var Version = "2.2.16"

const repositoriesEndpoint = "artifactory/api/repositories/"

type ArtClient struct {
	ArtOld *artifactoryold.Artifactory
	ArtNew *jfroghttpclient.JfrogHttpClient
	Xray   *xray.Xray
	Resty  *resty.Client
}

// Provider Artifactory provider that supports configuration via username+password or a token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_URL", func() (interface{}, error) {
					return "http://localhost:8082", nil
				}),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_USERNAME", func() (interface{}, error) {
					return "admin", nil
				}),
				ConflictsWith: []string{"access_token", "api_key"},
				ValidateFunc:  validation.StringIsNotEmpty,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_PASSWORD", func() (interface{}, error) {
					return "password", nil
				}),
				ConflictsWith: []string{"access_token", "api_key"},
				ValidateFunc:  validation.StringIsNotEmpty,
			},
			"api_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_API_KEY", nil),
				ConflictsWith: []string{"username", "access_token", "password"},
				ValidateFunc:  validation.StringIsNotEmpty,
			},
			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_ACCESS_TOKEN", nil),
				ConflictsWith: []string{"api_key", "password"},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"artifactory_local_repository":          resourceArtifactoryLocalRepository(),
			"artifactory_remote_repository":         resourceArtifactoryRemoteRepository(),
			"artifactory_virtual_repository":        resourceArtifactoryVirtualRepository(),
			"artifactory_group":                     resourceArtifactoryGroup(),
			"artifactory_user":                      resourceArtifactoryUser(),
			"artifactory_permission_target":         resourceArtifactoryPermissionTarget(),
			"artifactory_replication_config":        resourceArtifactoryReplicationConfig(),
			"artifactory_single_replication_config": resourceArtifactorySingleReplicationConfig(),
			"artifactory_certificate":               resourceArtifactoryCertificate(),
			"artifactory_api_key":                   resourceArtifactoryApiKey(),
			"artifactory_access_token":              resourceArtifactoryAccessToken(),
			// Deprecated. Remove in V3
			"artifactory_permission_targets": resourceArtifactoryPermissionTargets(),
			// Xray resources
			"artifactory_xray_policy": resourceXrayPolicy(),
			"artifactory_xray_watch":  resourceXrayWatch(),
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
			return fmt.Errorf("request failure %d: %s\n", response.StatusCode(), string(response.Body()[:]))
		}
		return nil
	}).
		SetHeader("content-type", "application/json").
		SetHeader("accept", "*/*").
		SetHeader("user-agent", "jfrog/terraform-provider-artifactory:"+Version)
	restyBase.DisableWarn = true
	return restyBase, nil

}

func addAuthToResty(client *resty.Client, username, password, apiKey, accessToken string) (*resty.Client, error) {

	if username != "" && password != "" {
		return client.SetBasicAuth(username, password), nil
	}
	if apiKey != "" {
		return client.SetHeader("X-JFrog-Art-Api", apiKey), nil
	}
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
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

	rt := &ArtClient{
		ArtOld: nil,
		ArtNew: nil,
		Xray:   nil,
		Resty:  restyBase,
	}

	return rt, nil

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
