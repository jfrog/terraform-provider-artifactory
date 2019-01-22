package artifactory

import (
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/pkg/artifactory"
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
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_PASSWORD", nil),
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARTIFACTORY_TOKEN", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"artifactory_local_repository":   resourceArtifactoryLocalRepository(),
			"artifactory_remote_repository":  resourceArtifactoryRemoteRepository(),
			"artifactory_virtual_repository": resourceArtifactoryVirtualRepository(),
			"artifactory_group":              resourceArtifactoryGroup(),
			"artifactory_user":               resourceArtifactoryUser(),
			"artifactory_permission_targets": resourceArtifactoryPermissionTargets(),
			"artifactory_replication_config": resourceArtifactoryReplicationConfig(),
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
	token := d.Get("token").(string)

	var client *http.Client
	if username != "" && password != "" {
		tp := artifactory.BasicAuthTransport{
			Username: username,
			Password: password,
		}
		client = tp.Client()
	} else if token != "" {
		tp := &artifactory.TokenAuthTransport{
			Token: token,
		}
		client = tp.Client()
	} else {
		return nil, fmt.Errorf("either [username, password] or [token] must be set to use provider")
	}

	return artifactory.NewClient(d.Get("url").(string), client)
}
