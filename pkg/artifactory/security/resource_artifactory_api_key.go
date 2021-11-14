package security

import (
	"fmt"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"strconv"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ApiKey struct {
	ApiKey string `json:"apiKey"`
}

const apiKeyEndpoint = "artifactory/api/security/apiKey"

func ResourceArtifactoryApiKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiKeyCreate,
		Read:   resourceApiKeyRead,
		Delete: apiKeyRevoke,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func packApiKey(apiKey string, d *schema.ResourceData) error {

	setValue := util.MkLens(d)

	errors := setValue("api_key", apiKey)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack api key %q", errors)
	}

	return nil
}

func resourceApiKeyCreate(d *schema.ResourceData, m interface{}) error {
	data := make(map[string]string)

	_, err := m.(*resty.Client).R().SetResult(&data).Post(apiKeyEndpoint)
	if err != nil {
		return err
	}

	if apiKey, ok := data["apiKey"]; ok {
		d.SetId(strconv.Itoa(schema.HashString(apiKey)))
		return resourceApiKeyRead(d, m)
	}
	return fmt.Errorf("received no error when creating apikey, but also got no apikey")
}

func resourceApiKeyRead(d *schema.ResourceData, m interface{}) error {
	data := make(map[string]string)
	_, err := m.(*resty.Client).R().SetResult(&data).Get(apiKeyEndpoint)
	if err != nil {
		return err
	}
	key := data["apiKey"]
	if key == "" {
		d.SetId("")
		return nil
	}
	return packApiKey(key, d)
}

func apiKeyRevoke(_ *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete(apiKeyEndpoint)
	return err
}
