package xray

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Version for some reason isn't getting updated by the linker
var Version = "0.0.1"

// Provider Xray provider that supports configuration via username+password or a token
// Supported resources are policies and watches
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"JFROG_URL", "XRAY_URL"},
					"http://localhost:8081"),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description:  "URL of Artifactory. This can also be sourced from the `XRAY_URL` or `JFROG_URL` environment variable. Default to 'http://localhost:8081' if not set.",
			},
			"access_token": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"XRAY_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN"}, ""),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description:      "This is a bearer token that can be given to you by your admin under `Identity and Access`",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			// Xray resources
			"xray_security_policy": resourceXraySecurityPolicyV2(),
			"xray_license_policy":  resourceXrayLicensePolicyV2(),
			"xray_watch":           resourceXrayWatch(),
			"xray_settings":        resourceXraySettings(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.13+compatible"
		}
		configuration, err := providerConfigure(data, terraformVersion)
		return configuration, diag.FromErr(err)
	}

	return p
}

func buildResty(URL string) (*resty.Client, error) {

	u, err := url.ParseRequestURI(URL)

	if err != nil {
		return nil, err
	}
	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	restyBase := resty.New().SetHostURL(baseUrl).OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		if response == nil {
			return fmt.Errorf("no response found")
		}
		if response.StatusCode() >= http.StatusBadRequest {
			return fmt.Errorf("\n%d %s %s\n%s", response.StatusCode(), response.Request.Method, response.Request.URL, string(response.Body()[:]))
		}
		return nil
	}).
		SetHeader("content-type", "application/json").
		SetHeader("accept", "*/*").
		SetHeader("user-agent", "jfrog/terraform-provider-xray:"+Version).
		SetRetryCount(5)
	restyBase.DisableWarn = true
	return restyBase, nil

}

func addAuthToResty(client *resty.Client, accessToken string) (*resty.Client, error) {
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
	}
	return nil, fmt.Errorf("no authentication details supplied")
}

// Creates the client for artifactory, will use token auth
func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	URL, ok := d.GetOk("url")
	if URL == nil || URL == "" || !ok {
		return nil, fmt.Errorf("you must supply a URL")
	}

	restyBase, err := buildResty(URL.(string))
	if err != nil {
		return nil, err
	}
	accessToken := d.Get("access_token").(string)

	restyBase, err = addAuthToResty(restyBase, accessToken)
	if err != nil {
		return nil, err
	}

	_, err = sendUsageRepo(restyBase, terraformVersion)
	if err != nil {
		return nil, err
	}

	err = checkArtifactoryLicense(restyBase)
	if err != nil {
		return nil, err
	}

	return restyBase, nil

}

func sendUsageRepo(restyBase *resty.Client, terraformVersion string) (interface{}, error) {
	type Feature struct {
		FeatureId string `json:"featureId"`
	}
	type UsageStruct struct {
		ProductId string    `json:"productId"`
		Features  []Feature `json:"features"`
	}
	_, err := restyBase.R().SetBody(UsageStruct{
		"terraform-provider-xray/" + Version,
		[]Feature{
			{FeatureId: "Partner/ACC-007450"},
			{FeatureId: "Terraform/" + terraformVersion},
		},
	}).Post("artifactory/api/system/usage")

	if err != nil {
		return nil, fmt.Errorf("unable to report usage %s", err)
	}
	return nil, nil
}

func checkArtifactoryLicense(client *resty.Client) error {

	type License struct {
		Type string `json:"type"`
	}

	type LicensesWrapper struct {
		License
		Licenses []License `json:"licenses"` // HA licenses returns as an array instead
	}

	licensesWrapper := LicensesWrapper{}
	_, err := client.R().
		SetResult(&licensesWrapper).
		Get("/artifactory/api/system/license")

	if err != nil {
		return fmt.Errorf("Failed to check for license. %s", err)
	}

	var licenseType string
	if len(licensesWrapper.Licenses) > 0 {
		licenseType = licensesWrapper.Licenses[0].Type
	} else {
		licenseType = licensesWrapper.Type
	}

	if matched, _ := regexp.MatchString(`Enterprise`, licenseType); !matched {
		return fmt.Errorf("Xray requires Enterprise license!")
	}

	return nil
}
