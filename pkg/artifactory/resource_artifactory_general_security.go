package artifactory

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		Update: resourceGeneralSecurityUpdate,
		Create: resourceGeneralSecurityUpdate,
		Delete: resourceGeneralSecurityDelete,
		Read:   resourceGeneralSecurityRead,

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

func resourceGeneralSecurityRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtNew
	serviceDetails := c.GetConfig().GetServiceDetails()
	httpClientDetails := serviceDetails.CreateHttpClientDetails()

	generalSettings := GeneralSettings{}

	_, body, _, err := c.Client().SendGet(fmt.Sprintf("%sapi/securityconfig", serviceDetails.GetUrl()), false, &httpClientDetails)
	if err != nil {
		return fmt.Errorf("failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
	}

	err = json.Unmarshal(body, &generalSettings)
	if err != nil {
		return fmt.Errorf("failed to unmarshal general security settings during Read")
	}

	s := GeneralSecurity{GeneralSettings: generalSettings}
	return packGeneralSecurity(&s, d)
}

func resourceGeneralSecurityUpdate(d *schema.ResourceData, m interface{}) error {
	unpacked := unpackGeneralSecurity(d)
	content, err := yaml.Marshal(&unpacked)

	if err != nil {
		return fmt.Errorf("failed to marshal general security settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return fmt.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one general security settings resource, using same id
	d.SetId("security")
	return resourceGeneralSecurityRead(d, m)

	return nil
}

func resourceGeneralSecurityDelete(d *schema.ResourceData, m interface{}) error {
	var content = `
security:
  anonAccessEnabled: false
`

	err := sendConfigurationPatch([]byte(content), m)
	if err != nil {
		return fmt.Errorf("failed to send PATCH request to Artifactory during Delete")
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

func packGeneralSecurity(s *GeneralSecurity, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("enable_anonymous_access", s.GeneralSettings.AnonAccessEnabled))

	if hasErr {
		return fmt.Errorf("failed to pack general security settings")
	}

	return nil
}
