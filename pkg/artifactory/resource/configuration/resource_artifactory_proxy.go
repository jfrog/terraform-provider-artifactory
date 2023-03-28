package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Key               string `xml:"key" yaml:"-"`
	Host              string `xml:"host" yaml:"host"`
	Port              int    `xml:"port" yaml:"port"`
	Username          string `xml:"username" yaml:"username"`
	Password          string `xml:"password" yaml:"password"`
	NtHost            string `xml:"ntHost" yaml:"ntHost"`
	NtDomain          string `xml:"domain" yaml:"domain"`
	PlatformDefault   bool   `xml:"platformDefault" yaml:"platformDefault"`
	RedirectedToHosts string `xml:"redirectedToHosts" yaml:"redirectedToHosts"`
	Services          string `xml:"services" yaml:"services"`
}

func (p Proxy) Id() string {
	return p.Key
}

type Proxies struct {
	Proxies []Proxy `xml:"proxies>proxy" yaml:"proxy"`
}

func ResourceArtifactoryProxy() *schema.Resource {
	var proxySchema = map[string]*schema.Schema{
		"key": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The unique ID of the proxy.",
		},
		"host": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The name of the proxy host.",
		},
		"port": {
			Type:             schema.TypeInt,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			Description:      "The proxy port number.",
		},
		"username": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The proxy username when authentication credentials are required.",
		},
		"password": {
			Type:             schema.TypeString,
			Optional:         true,
			Sensitive:        true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The proxy password when authentication credentials are required.",
		},
		"nt_host": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The computer name of the machine (the machine connecting to the NTLM proxy).",
		},
		"nt_domain": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "The proxy domain/realm name.",
		},
		"platform_default": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, this proxy will be the default proxy for new remote repositories and for internal HTTP requests issued by Artifactory. Will also be used as proxy for all other services in the platform (for example: Xray, Distribution, etc).",
		},
		"redirect_to_hosts": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "An optional list of host names to which this proxy may redirect requests. The credentials defined for the proxy are reused by requests redirected to all of these hosts.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"services": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "An optional list of services names to which this proxy be the default of. The options are jfrt, jfmc, jfxr, jfds",
			Elem: &schema.Schema{
				Type: schema.TypeString,
				Set:  schema.HashString,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{
							"jfrt",
							"jfmc",
							"jfxr",
							"jfds",
						},
						false,
					),
				),
			},
		},
	}

	var unpackProxy = func(s *schema.ResourceData) Proxy {
		d := &util.ResourceData{ResourceData: s}
		return Proxy{
			Key:               d.GetString("key", false),
			Host:              d.GetString("host", false),
			Port:              d.GetInt("port", false),
			Username:          d.GetString("username", false),
			Password:          d.GetString("password", false),
			NtHost:            d.GetString("nt_host", false),
			NtDomain:          d.GetString("nt_domain", false),
			PlatformDefault:   d.GetBool("platform_default", false),
			RedirectedToHosts: strings.Join(d.GetSet("redirect_to_hosts"), ","),
			Services:          strings.Join(d.GetSet("services"), ","),
		}
	}

	var packProxy = func(p *Proxy, d *schema.ResourceData) diag.Diagnostics {
		setValue := util.MkLens(d)

		setValue("key", p.Key)
		setValue("host", p.Host)
		setValue("port", p.Port)
		setValue("username", p.Username)
		setValue("nt_host", p.NtHost)
		setValue("nt_domain", p.NtDomain)
		errors := setValue("platform_default", p.PlatformDefault)
		if p.RedirectedToHosts != "" {
			errors = setValue("redirect_to_hosts", strings.Split(p.RedirectedToHosts, ","))
		}
		if p.Services != "" {
			errors = setValue("services", strings.Split(p.Services, ","))
		}

		if errors != nil && len(errors) > 0 {
			return diag.Errorf("failed to pack proxy %q", errors)
		}

		return nil
	}

	var resourceProxyRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		data := &util.ResourceData{ResourceData: d}
		key := data.GetString("key", false)

		proxiesConfig := Proxies{}
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&proxiesConfig).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedProxyConfig := FindConfigurationById[Proxy](proxiesConfig.Proxies, key)
		if matchedProxyConfig == nil {
			d.SetId("")
			return nil
		}

		return packProxy(matchedProxyConfig, d)
	}

	var resourceProxyUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedProxy := unpackProxy(d)

		///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
		//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
		//GET call structure has "propertySets -> propertySet -> Array of property sets".
		//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
		//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
		//*/
		var body = map[string]map[string]Proxy{
			"proxies": {
				unpackedProxy.Key: unpackedProxy,
			},
		}

		content, err := yaml.Marshal(&body)
		if err != nil {
			return diag.Errorf("failed to marshal proxy during Update")
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Update")
		}

		d.SetId(unpackedProxy.Key)
		return resourceProxyRead(ctx, d, m)
	}

	var resourceProxyDelete = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		proxiesConfig := &Proxies{}

		response, err := m.(util.ProvderMetadata).Client.R().SetResult(&proxiesConfig).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return diag.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		matchedProxyConfig := FindConfigurationById[Proxy](proxiesConfig.Proxies, d.Id())
		if matchedProxyConfig == nil {
			return diag.Errorf("No proxy found for '%s'", d.Id())
		}

		var body = map[string]map[string]string{
			"proxies": {
				matchedProxyConfig.Key: "~",
			},
		}

		content, err := yaml.Marshal(&body)
		if err != nil {
			return diag.Errorf("failed to marshal proxy during Delete")
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
		}

		d.SetId("")

		return nil
	}

	var verifyCrossDependentValues = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		platformDefault := diff.Get("platform_default").(bool)
		services := diff.Get("services").(*schema.Set).List()

		if platformDefault && len(services) > 0 {
			return fmt.Errorf("services cannot be set when platform_default is true")
		}

		return nil
	}

	return &schema.Resource{
		UpdateContext: resourceProxyUpdate,
		CreateContext: resourceProxyUpdate,
		DeleteContext: resourceProxyDelete,
		ReadContext:   resourceProxyRead,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("key", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema:        proxySchema,
		CustomizeDiff: verifyCrossDependentValues,
		Description:   "Provides an Artifactory Proxy resource. This resource configuration is only available for self-hosted instance. It corresponds to 'proxies' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}
