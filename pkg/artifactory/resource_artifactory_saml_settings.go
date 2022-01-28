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
	UseEncryptedAssertion     bool   `yaml:"useEncryptedAssertion" json:"useEncryptedAssertion"`
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `(Optional) Enable SAML SSO.  Default value is "true".`,
			},
			"certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `(Optional) SAML certificate that contains the public key for the IdP service provider.  Used by Artifactory to verify sign-in requests. Default value is "".`,
			},
			"email_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `(Optional) Name of the attribute in the SAML response from the IdP that contains the user's email. Default value is "".`,
			},
			"group_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `(Optional) Name of the attribute in the SAML response from the IdP that contains the user's group memberships. Default value is "".`,
			},
			"login_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `(Required) Service provider login url configured on the IdP.`,
			},
			"logout_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `(Required) Service provider logout url, or where to redirect after user logs out.`,
			},
			"no_auto_user_creation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `(Optional) When automatic user creation is off, authenticated users are not automatically created inside Artifactory. Instead, for every request from an SSO user, the user is temporarily associated with default groups (if such groups are defined), and the permissions for these groups apply. Without auto-user creation, you must manually create the user inside Artifactory to manage user permissions not attached to their default groups. Default value is "false".`,
			},
			"service_provider_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `(Required) The SAML service provider name. This should be a URI that is also known as the entityID, providerID, or entity identity.`,
			},
			"allow_user_to_access_profile": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `(Optional) Allow persisted users to access their profile.  Default value is "true".`,
			},
			"auto_redirect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `(Optional) Auto redirect to login through the IdP when clicking on Artifactory's login link.  Default value is "false".`,
			},
			"sync_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `(Optional) Associate user with Artifactory groups based on the "group_attribute" provided in the SAML response from the identity provider.  Default value is "false".`,
			},
			"verify_audience_restriction": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `(Optional) Enable "audience", or who the SAML assertion is intended for.  Ensures that the correct service provider intended for Artifactory is used on the IdP. Default value is "true".`,
			},
			"use_encrypted_assertion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `(Optional) When set, an X.509 public certificate will be created by Artifactory. Download this certificate and upload it to your IDP and choose your own encryption algorithm. This process will let you encrypt the assertion section in your SAML response. Default value is "false".`,
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
		UseEncryptedAssertion:     d.getBool("use_encrypted_assertion", false),
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
	setValue("use_encrypted_assertion", s.Saml.Settings.UseEncryptedAssertion)
	errors := setValue("verify_audience_restriction", s.Saml.Settings.VerifyAudienceRestriction)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack saml settings %q", errors)
	}

	return nil
}
