package artifactory

import (
	"context"
	"fmt"
	"net/http"

	artifactoryold "github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/version"
	artifactorynew "github.com/jfrog/jfrog-client-go/artifactory"
	artifactoryauth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/usage"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/xray"
	xrayauth "github.com/jfrog/jfrog-client-go/xray/auth"
)

var ProviderVersion = "2.1.0"

type ArtClient struct {
	ArtOld     *artifactoryold.Artifactory
	ArtNew     *artifactorynew.ArtifactoryServicesManager
	XrayClient *xray.XrayServicesManager
}

// Artifactory Provider that supports configuration via username+password or a token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_URL", nil),
			},
			"xray_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("XRAY_URL", nil),
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
			"artifactory_watch":                     resourceArtifactoryWatch(),
			// Deprecated. Remove in V3
			"artifactory_permission_targets": resourceArtifactoryPermissionTargets(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"artifactory_file":     dataSourceArtifactoryFile(),
			"artifactory_fileinfo": dataSourceArtifactoryFileInfo(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	if d.Get("url") == nil {
		return nil, fmt.Errorf("url cannot be nil")
	}

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	log.SetLogger(log.NewLogger(log.INFO, nil))

	var client *http.Client
	details := artifactoryauth.NewArtifactoryDetails()

	url := d.Get("url").(string)
	if url[len(url)-1] != '/' {
		url += "/"
	}
	details.SetUrl(url)

	if username != "" && password != "" {
		details.SetUser(username)
		details.SetPassword(password)
		tp := transport.BasicAuth{
			Username: username,
			Password: password,
		}
		client = tp.Client()
	} else if apiKey != "" {
		details.SetApiKey(apiKey)
		tp := &transport.ApiKeyAuth{
			ApiKey: apiKey,
		}
		client = tp.Client()
	} else if accessToken != "" {
		details.SetAccessToken(accessToken)
		tp := &transport.AccessTokenAuth{
			AccessToken: accessToken,
		}
		client = tp.Client()
	} else {
		return nil, fmt.Errorf("either [username, password] or [api_key] or [access_token] must be set to use provider")
	}

	config, err := config.NewConfigBuilder().
		SetServiceDetails(details).
		SetDryRun(false).
		Build()

	if err != nil {
		return nil, err
	}

	rtold, err := artifactoryold.NewClient(d.Get("url").(string), client)

	if err != nil {
		return nil, err
	}

	rtnew, err := artifactorynew.New(&details, config)

	if err != nil {
		return nil, err
	} else if _, resp, err := rtold.V1.System.Ping(context.Background()); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to ping server. Got %d", resp.StatusCode)
	} else if _, err := rtnew.Ping(); err != nil {
		return nil, err
	}

	productid := "terraform-provider-artifactory/" + ProviderVersion
	commandid := "Terraform/" + version.Version
	usage.SendReportUsage(productid, commandid, rtnew)

	var xrayClient *xray.XrayServicesManager
	if d.Get("xray_url") != nil && d.Get("xray_url").(string) != "" {
		xrayClient, err = createXrayClient(d)
		if err != nil {
			return nil, err
		}
	}

	rt := &ArtClient{
		ArtOld:     rtold,
		ArtNew:     &rtnew,
		XrayClient: xrayClient,
	}

	return rt, nil
}

func createXrayClient(d *schema.ResourceData) (*xray.XrayServicesManager, error) {
	details := xrayauth.NewXrayDetails()

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	url := d.Get("xray_url").(string)
	if url[len(url)-1] != '/' {
		url += "/"
	}
	details.SetUrl(url)

	if username != "" && password != "" {
		details.SetUser(username)
		details.SetPassword(password)
	} else if apiKey != "" {
		details.SetApiKey(apiKey)
	} else if accessToken != "" {
		details.SetAccessToken(accessToken)
	} else {
		return nil, fmt.Errorf("either [username, password] or [api_key] or [access_token] must be set to use provider")
	}

	config, err := config.NewConfigBuilder().
		SetServiceDetails(details).
		SetDryRun(false).
		Build()

	if err != nil {
		return nil, err
	}

	cd := auth.ServiceDetails(details)
	xrayClient, err := xray.New(&cd, config)

	return xrayClient, nil
}
