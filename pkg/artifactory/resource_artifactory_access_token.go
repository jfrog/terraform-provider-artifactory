package artifactory

import (
	"context"
	"net/http"
	"strconv"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArtifactoryAcessToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessTokenCreate,
		Read:   resourceAccessTokenRead,
		Delete: resourceAccessTokenDelete,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"expires_in": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
				ForceNew: true,
			},
			"refreshable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"access_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"refresh_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func unmarshalToken(s *schema.ResourceData) *v1.AccessTokenOptions {
	d := &ResourceData{s}

	accessTokenOption := new(v1.AccessTokenOptions)

	accessTokenOption.Username = d.getStringRef("username", false)
	accessTokenOption.Scope = d.getStringRef("scope", false)
	accessTokenOption.ExpiresIn = d.getIntRef("expires_in", false)
	refreshable := d.getBoolRef("refreshable", false)
	accessTokenOption.Refreshable = artifactory.String(strconv.FormatBool(*refreshable))
	accessTokenOption.Audience = d.getStringRef("audience", false)

	return accessTokenOption
}

func resourceAccessTokenCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	accessTokenOption := unmarshalToken(d)

	token, _, err := c.V1.Security.CreateToken(context.Background(), accessTokenOption)
	if err != nil {
		return err
	}

	d.SetId(*token.AccessToken)
	d.Set("access_token", *token.AccessToken)
	d.Set("refresh_token", *token.RefreshToken)

	return nil
}

func resourceAccessTokenRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAccessTokenDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	_, resp, err := c.V1.Security.RevokeToken(context.Background(), v1.AccessTokenRevokeOptions{
		Token: d.Id(),
	})
	if err != nil {
		return err
	}	

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return nil
	}

	return err
}
