package artifactory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

type OauthSecurity struct {
	Oauth OauthSettingsWrapper `yaml:"security"`
}

type OauthSettingsWrapper struct {
	Settings OauthSettings `yaml:"oauthSettings"`
}

type OauthSettings struct {
	EnableIntegration        bool                             `yaml:"enableIntegration" json:"enabled"`
	PersistUsers             bool                             `yaml:"persistUsers" json:"persistUsers"`
	AllowUserToAccessProfile bool                             `yaml:"allowUserToAccessProfile" json:"allowUserToAccessProfile"`
	OauthProvidersSettings   map[string]OauthProviderSettings `yaml:"oauthProvidersSettings"`
	AvailableTypes           []OauthType                      `json:"availableTypes"`
	Providers                []OauthProviderSettings          `json:"providers"`
}

type OauthProviderSettings struct {
	Name         string `json:"name"`
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	Type         string `yaml:"providerType" json:"providerType"`
	ClientId     string `yaml:"id" json:"id"`
	ClientSecret string `yaml:"secret" json:"secret"`
	ApiUrl       string `yaml:"apiUrl" json:"apiUrl"`
	AuthUrl      string `yaml:"authUrl" json:"authUrl"`
	TokenUrl     string `yaml:"tokenUrl" json:"tokenUrl"`
}

type OauthType struct {
	DisplayName     string
	Type            string
	MandatoryFields []string
	FieldHolders    []string
	FieldValues     []string
}

func resourceArtifactoryOauthSettings() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceOauthSettingsUpdate,
		CreateContext: resourceOauthSettingsUpdate,
		DeleteContext: resourceOauthSettingsDelete,
		ReadContext:   resourceOauthSettingsRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"persist_users": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_user_to_access_profile": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"oauth_provider": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"client_secret": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"api_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"auth_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"token_url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceOauthSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*ArtClient).Resty

	oauthSettings := OauthSettings{}
	 _, err := c.R().SetResult(&oauthSettings).Get("artifactory/api/oauth")

	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/oauth during Read")
	}

	s := OauthSecurity{OauthSettingsWrapper{Settings: oauthSettings}}
	s.Oauth.Settings.OauthProvidersSettings = make(map[string]OauthProviderSettings)

	for _, provider := range s.Oauth.Settings.Providers {
		s.Oauth.Settings.OauthProvidersSettings[provider.Name] = provider
	}

	// remove resource from state if no providers are found
	if len(s.Oauth.Settings.Providers) == 0 {
		d.SetId("")
	}

	packDiag := packOauthSecurity(&s, d)
	if packDiag != nil {
		return packDiag
	}

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Usage of Undocumented Artifactory API Endpoints",
		Detail:   "The artifactory_oauth_settings resource uses endpoints that are undocumented and may not work with SaaS environments, or may change without notice.",
	}}
}

func resourceOauthSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpacked := unpackOauthSecurity(d)
	content, err := yaml.Marshal(&unpacked)

	if err != nil {
		return diag.Errorf("failed to marshal oauth settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one oauth settings resource, using same id
	d.SetId("oauth_settings")
	return resourceOauthSettingsRead(ctx, d, m)
}

func resourceOauthSettingsDelete(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	var content = `
security:
  oauthSettings: ~
`

	err := sendConfigurationPatch([]byte(content), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
	}
	return nil
}

func unpackOauthSecurity(s *schema.ResourceData) *OauthSecurity {
	d := &ResourceData{s}

	security := new(OauthSecurity)

	settings := OauthSettings{
		EnableIntegration:        d.getBool("enable", false),
		PersistUsers:             d.getBool("persist_users", false),
		AllowUserToAccessProfile: d.getBool("allow_user_to_access_profile", false),
	}

	if v, ok := d.GetOkExists("oauth_provider"); ok {
		oauthProviderSettings := map[string]OauthProviderSettings{}

		for _, m := range v.(*schema.Set).List() {
			o := m.(map[string]interface{})

			oauthProviderSettings[o["name"].(string)] = OauthProviderSettings{
				Name:         o["name"].(string),
				Enabled:      o["enabled"].(bool),
				Type:         o["type"].(string),
				ClientId:     o["client_id"].(string),
				ClientSecret: o["client_secret"].(string),
				ApiUrl:       o["api_url"].(string),
				AuthUrl:      o["auth_url"].(string),
				TokenUrl:     o["token_url"].(string),
			}
		}

		settings.OauthProvidersSettings = oauthProviderSettings
		security.Oauth.Settings = settings
	}
	return security
}

func packOauthSecurity(s *OauthSecurity, d *schema.ResourceData) diag.Diagnostics {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("enable", s.Oauth.Settings.EnableIntegration))
	logErr(d.Set("persist_users", s.Oauth.Settings.PersistUsers))
	logErr(d.Set("allow_user_to_access_profile", s.Oauth.Settings.AllowUserToAccessProfile))

	settings := make([]interface{}, 0)

	for name, setting := range s.Oauth.Settings.OauthProvidersSettings {
		providerSetting := map[string]interface{}{
			"name":          name,
			"enabled":       setting.Enabled,
			"type":          setting.Type,
			"client_id":     setting.ClientId,
			"client_secret": setting.ClientSecret,
			"api_url":       setting.ApiUrl,
			"auth_url":      setting.AuthUrl,
			"token_url":     setting.TokenUrl,
		}

		settings = append(settings, providerSetting)
	}

	if hasErr {
		return diag.Errorf("failed to pack oauth settings")
	}
	return nil
}
