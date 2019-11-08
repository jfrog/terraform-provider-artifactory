package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

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
			"token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("ARTIFACTORY_TOKEN", nil),
				ConflictsWith: []string{"api_key"},
				Deprecated:    "Since v1.5. Renamed to api_key",
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
			// Deprecated. Remove in V3
			"artifactory_permission_targets": resourceArtifactoryPermissionTargets(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"artifactory_file": datasourceArtifactoryFile(),
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

	// Deprecated
	token := d.Get("token").(string)

	var client *http.Client
	if username != "" && password != "" {
		tp := transport.BasicAuth{
			Username: username,
			Password: password,
		}
		client = tp.Client()
	} else if apiKey != "" {
		tp := &transport.ApiKeyAuth{
			ApiKey: apiKey,
		}
		client = tp.Client()
	} else if accessToken != "" {
		tp := &transport.AccessTokenAuth{
			AccessToken: accessToken,
		}
		client = tp.Client()
	} else if token != "" {
		tp := &transport.ApiKeyAuth{
			ApiKey: token,
		}
		client = tp.Client()
	} else {
		return nil, fmt.Errorf("either [username, password] or [api_key] or [access_token] must be set to use provider")
	}

	rt, err := artifactory.NewClient(d.Get("url").(string), client)

	if err != nil {
		return nil, err
	} else if _, resp, err := rt.V1.System.Ping(context.Background()); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to ping server. Got %d", resp.StatusCode)
	}

	return rt, nil
}
