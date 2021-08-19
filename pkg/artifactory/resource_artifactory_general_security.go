package artifactory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

type GeneralSecurity struct {
	GeneralSettings `yaml:"security" json:"security"`
}

type GeneralSettings struct {
	AnonAccessEnabled bool `yaml:"anonAccessEnabled" json:"anonAccessEnabled"`
}

func resourceArtifactoryGeneralSecurity() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceGeneralSecurityUpdate,
		CreateContext: resourceGeneralSecurityUpdate,
		DeleteContext: resourceGeneralSecurityDelete,
		ReadContext:   resourceGeneralSecurityRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable_anonymous_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceGeneralSecurityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*ArtClient).ArtNew
	serviceDetails := c.GetConfig().GetServiceDetails()
	httpClientDetails := serviceDetails.CreateHttpClientDetails()

	generalSettings := GeneralSettings{}

	_, body, _, err := c.Client().SendGet(fmt.Sprintf("%sapi/securityconfig", serviceDetails.GetUrl()), false, &httpClientDetails)
	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read.  If you are using the SaaS offering of Artifactory this feature is not supported")
	}

	err = json.Unmarshal(body, &generalSettings)
	if err != nil {
		return diag.Errorf("failed to unmarshal general security settings during Read")
	}

	s := GeneralSecurity{GeneralSettings: generalSettings}
	packDiag := packGeneralSecurity(&s, d)

	if packDiag != nil {
		return packDiag
	}

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "the general security resource uses undocumented API endpoints",
		Detail:   "the general security resource uses Artifactory endpoints that are undocumented and do not exist in the SaaS version",
	}}
}

func resourceGeneralSecurityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpacked := unpackGeneralSecurity(d)
	content, err := yaml.Marshal(&unpacked)

	if err != nil {
		return diag.Errorf("failed to marshal general security settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one general security settings resource, using same id
	d.SetId("security")
	return resourceGeneralSecurityRead(ctx, d, m)

	return nil
}

func resourceGeneralSecurityDelete(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	var content = `
security:
  anonAccessEnabled: false
`

	err := sendConfigurationPatch([]byte(content), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
	}

	return nil
}

func unpackGeneralSecurity(s *schema.ResourceData) *GeneralSecurity {
	d := &ResourceData{s}
	security := *new(GeneralSecurity)

	settings := GeneralSettings{
		AnonAccessEnabled: *d.getBoolRef("enable_anonymous_access", false),
	}

	security.GeneralSettings = settings
	return &security
}

func packGeneralSecurity(s *GeneralSecurity, d *schema.ResourceData) diag.Diagnostics {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("enable_anonymous_access", s.GeneralSettings.AnonAccessEnabled))

	if hasErr {
		return diag.Errorf("failed to pack general security settings")
	}

	return nil
}
