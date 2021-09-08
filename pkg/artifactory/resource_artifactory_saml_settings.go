package artifactory

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

type SamlSecurity struct {
	Saml SamlSettingsWrapper `yaml:"security"`
}

type SamlSettingsWrapper struct {
	Settings SamlSettings `yaml:"samlSettings"`
}

type SamlSettings struct {
	EnableIntegration         bool   `yaml:"enableIntegration" json:"enableIntegration"`
	Certificate               string `yaml:"certificate" json:"certificate"`
	EmailAttribute            string `yaml:"emailAttribute" json:"emailAttribute"`
	GroupAttribute            string `yaml:"groupAttribute" json:"groupAttribute"`
	LoginUrl                  string `yaml:"loginUrl" json:"loginUrl"`
	LogoutUrl                 string `yaml:"logoutUrl" json:"logoutUrl"`
	NoAutoUserCreation        bool   `yaml:"noAutoUserCreation" json:"noAutoUserCreation"`
	ServiceProviderName       string `yaml:"serviceProviderName" json:"serviceProviderName"`
	AllowUserToAccessProfile  bool   `yaml:"allowUserToAccessProfile" json:"allowUserToAccessProfile"`
	AutoRedirect              bool   `yaml:"autoRedirect" json:"autoRedirect"`
	SyncGroups                bool   `yaml:"syncGroups" json:"syncGroups"`
	VerifyAudienceRestriction bool   `yaml:"verifyAudienceRestriction" json:"verifyAudienceRestriction"`
}

func resourceArtifactorySamlSettings() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceSamlSettingsUpdate,
		CreateContext: resourceSamlSettingsUpdate,
		DeleteContext: resourceSamlSettingsDelete,
		ReadContext:   resourceSamlSettingsRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"logout_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"no_auto_user_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"service_provider_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"allow_user_to_access_profile": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"auto_redirect": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sync_groups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"verify_audience_restriction": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceSamlSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*resty.Client)

	samlSettings := SamlSettings{}

	_, err := c.R().SetResult(&samlSettings).Get("artifactory/api/saml/config")
	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/saml/config during Read")
	}

	s := SamlSecurity{SamlSettingsWrapper{Settings: samlSettings}}

	packDiag := packSamlSecurity(&s, d)
	if packDiag != nil {
		return packDiag
	}

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Usage of Undocumented Artifactory API Endpoints",
		Detail:   "The artifactory_saml_settings resource uses endpoints that are undocumented and may not work with SaaS environments, or may change without notice.",
	}}
}

func resourceSamlSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpacked := unpackSamlSecurity(d)
	content, err := yaml.Marshal(&unpacked)

	if err != nil {
		return diag.Errorf("failed to marshal saml settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one saml settings resource, using same id
	d.SetId("saml_settings")
	return resourceSamlSettingsRead(ctx, d, m)
}

func resourceSamlSettingsDelete(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	var content = `
security:
  samlSettings: ~
`

	err := sendConfigurationPatch([]byte(content), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
	}

	return nil
}

func unpackSamlSecurity(s *schema.ResourceData) *SamlSecurity {
	d := &ResourceData{s}
	security := *new(SamlSecurity)

	settings := SamlSettings{
		EnableIntegration:         d.getBool("enable", false),
		Certificate:               d.getString("certificate", false),
		EmailAttribute:            d.getString("email_attribute", false),
		GroupAttribute:            d.getString("group_attribute", false),
		LoginUrl:                  d.getString("login_url", false),
		LogoutUrl:                 d.getString("logout_url", false),
		NoAutoUserCreation:        d.getBool("no_auto_user_creation", false),
		ServiceProviderName:       d.getString("service_provider_name", false),
		AllowUserToAccessProfile:  d.getBool("allow_user_to_access_profile", false),
		AutoRedirect:              d.getBool("auto_redirect", false),
		SyncGroups:                d.getBool("sync_groups", false),
		VerifyAudienceRestriction: d.getBool("verify_audience_restriction", false),
	}

	security.Saml.Settings = settings
	return &security
}

func packSamlSecurity(s *SamlSecurity, d *schema.ResourceData) diag.Diagnostics {
	setValue := mkLens(d)

	setValue("enable", s.Saml.Settings.EnableIntegration)
	setValue("certificate", s.Saml.Settings.Certificate)
	setValue("email_attribute", s.Saml.Settings.EmailAttribute)
	setValue("group_attribute", s.Saml.Settings.GroupAttribute)
	setValue("login_url", s.Saml.Settings.LoginUrl)
	setValue("logout_url", s.Saml.Settings.LogoutUrl)
	setValue("no_auto_user_creation", s.Saml.Settings.NoAutoUserCreation)
	setValue("service_provider_name", s.Saml.Settings.ServiceProviderName)
	setValue("allow_user_to_access_profile", s.Saml.Settings.AllowUserToAccessProfile)
	setValue("auto_redirect", s.Saml.Settings.AutoRedirect)
	setValue("sync_groups", s.Saml.Settings.SyncGroups)
	errors := setValue("verify_audience_restriction", s.Saml.Settings.VerifyAudienceRestriction)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack saml settings %q", errors)
	}

	return nil
}
