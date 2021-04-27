package artifactory

import (
	"fmt"
	artifactoryold "github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/jasonwbarnett/go-xray/xray"
	artifactorynew "github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/usage"
	auth2 "github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"net/url"
)

var repoTypeValidator = validation.StringInSlice([]string{
	"alpine",
	"bower",
	"cargo",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"p2",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
	"vcs",
	"yum",
}, false)

var ProviderVersion = "2.1.0"

type ArtClient struct {
	ArtOld *artifactoryold.Artifactory
	ArtNew *artifactorynew.ArtifactoryServicesManager
	Xray   *xray.Xray
}

// Artifactory Provider that supports configuration via username+password or a token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_URL", nil),
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_USERNAME", nil),
				ConflictsWith: []string{"access_token", "api_key"},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_PASSWORD", nil),
				ConflictsWith: []string{"access_token", "api_key"},
			},
			"api_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_API_KEY", nil),
				ConflictsWith: []string{"username", "access_token", "password"},
			},
			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_ACCESS_TOKEN", nil),
				ConflictsWith: []string{"username", "api_key", "password"},
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

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {

	if key, ok := d.GetOk("url"); key == nil || key == "" || !ok {
		return nil, fmt.Errorf("you must supply a URL")
	}

	log.SetLogger(log.NewLogger(log.INFO, nil))

	u, err := url.ParseRequestURI(d.Get("url").(string))

	if err != nil {
		return nil, err
	}
	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	artifactoryEndpoint := fmt.Sprintf("%s/artifactory/", baseUrl)

	client, details, err := buildClient(d)
	if err != nil {
		return nil, err
	}
	details.SetUrl(artifactoryEndpoint)

	cfg, err := config.NewConfigBuilder().
		SetServiceDetails(details).
		SetDryRun(false).
		Build()

	if err != nil {
		return nil, err
	}

	rtOld, err := artifactoryold.NewClient(artifactoryEndpoint, client)

	if err != nil {
		return nil, err
	}

	rtNew, err := artifactorynew.New(&details, cfg)

	if err != nil {
		return nil, err
	} else if _, err := rtNew.Ping(); err != nil {
		return nil, err
	}

	rtXray, err := xray.NewClient(fmt.Sprintf("%s/xray/", baseUrl), client)
	if err != nil {
		return nil, err
	}

	productId := "terraform-provider-artifactory/" + ProviderVersion
	commandId := "Terraform/" + terraformVersion
	if err = usage.SendReportUsage(productId, commandId, rtNew); err != nil {
		return nil, err
	}

	rt := &ArtClient{
		ArtOld: rtOld,
		ArtNew: rtNew,
		Xray:   rtXray,
	}

	return rt, nil
}

func buildClient(d *schema.ResourceData) (*http.Client, auth2.ServiceDetails, error) {
	details := auth.NewArtifactoryDetails()

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	if username != "" && password != "" {
		details.SetUser(username)
		details.SetPassword(password)
		tp := transport.BasicAuth{
			Username: username,
			Password: password,
		}
		return tp.Client(), details, nil
	} else if apiKey != "" {
		details.SetApiKey(apiKey)
		tp := &transport.ApiKeyAuth{
			ApiKey: apiKey,
		}
		return tp.Client(), details, nil
	} else if accessToken != "" {
		details.SetAccessToken(accessToken)
		tp := &transport.AccessTokenAuth{
			AccessToken: accessToken,
		}
		return tp.Client(), details, nil
	} else {
		return nil, nil, fmt.Errorf("either [username, password] or [api_key] or [access_token] must be set to use provider")
	}
}
