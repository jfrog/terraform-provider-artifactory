package artifactory

import (
	"fmt"

	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

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

	if token, ok := d.GetOkExists("token"); ok {
		tp := artifactory.TokenAuthTransport{
			Token: token.(string),
		}
		return artifactory.NewClient(d.Get("url").(string), tp.Client())
	} else if username, ok := d.GetOkExists("username"); !ok {
		return nil, fmt.Errorf("error: Missing token and username. One must be set")
	} else if password, ok := d.GetOkExists("password"); !ok {
		return nil, fmt.Errorf("error: Basic auth used but password not set")
	} else {
		tp := artifactory.BasicAuthTransport{
			Username: username.(string),
			Password: password.(string),
		}
		return artifactory.NewClient(d.Get("url").(string), tp.Client())
	}
}
