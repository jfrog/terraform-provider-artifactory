package security

import (
	"context"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
)

const ApiKeyEndpoint = "artifactory/api/security/apiKey"

type ApiKey struct {
	ApiKey            string `json:"apiKey"`
	BlockCreateApiKey bool   `json:"blockCreateApiKey"` // not used currently. may in future.
}

func ResourceArtifactoryApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiKeyCreate,
		ReadContext:   resourceApiKeyRead,
		DeleteContext: apiKeyRevoke,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func packApiKey(apiKey string, d *schema.ResourceData) diag.Diagnostics {

	setValue := util.MkLens(d)

	errors := setValue("api_key", apiKey)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack api key %q", errors)
	}

	return nil
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data := ApiKey{}

	_, err := m.(*resty.Client).R().SetResult(&data).Post(ApiKeyEndpoint)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(data.ApiKey) > 0 {
		d.SetId(strconv.Itoa(schema.HashString(data.ApiKey)))
		return resourceApiKeyRead(ctx, d, m)
	}
	return diag.Errorf("received no error when creating apikey, but also got no apikey")
}

func resourceApiKeyRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data := ApiKey{}
	_, err := m.(*resty.Client).R().SetResult(&data).Get(ApiKeyEndpoint)
	if err != nil {
		return diag.FromErr(err)
	}
	if data.ApiKey == "" {
		d.SetId("")
		return nil
	}
	return packApiKey(data.ApiKey, d)
}

func apiKeyRevoke(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(*resty.Client).R().Delete(ApiKeyEndpoint)
	return diag.FromErr(err)
}
