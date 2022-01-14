package artifactory

import (
	"context"
	"encoding/xml"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

type LdapSetting struct {
	Key                      string         `xml:"key" yaml:"key"`
	Enabled                  bool           `xml:"enabled" yaml:"enabled"`
	LdapUrl                  string         `xml:"ldapUrl" yaml:"ldapUrl"`
	UserDnPattern            string         `xml:"userDnPattern" yaml:"userDnPattern"`
	EmailAttribute           string         `xml:"emailAttribute" yaml:"emailAttribute"`
	AutoCreateUser           bool           `xml:"autoCreateUser" yaml:"autoCreateUser"`
	LdapPoisoningProtection  bool           `xml:"ldapPoisoningProtection" yaml:"ldapPoisoningProtection"`
	AllowUserToAccessProfile bool           `xml:"allowUserToAccessProfile" yaml:"allowUserToAccessProfile"`
	PagingSupportEnabled     bool           `xml:"pagingSupportEnabled" yaml:"pagingSupportEnabled"`
	Search                   LdapSearchType `xml:"search" yaml:"search"`
}

type LdapSearchType struct {
	SearchSubTree   bool   `yaml:"searchSubTree" xml:"searchSubTree"`
	SearchFilter    string `yaml:"searchFilter" xml:"searchFilter"`
	SearchBase      string `yaml:"searchBase" xml:"searchBase"`
	ManagerDn       string `yaml:"managerDn" xml:"managerDn"`
	ManagerPassword string `yaml:"managerPassword" xml:"managerPassword"`
}

type LdapSettings struct {
	LdapSettingArr []LdapSetting `yaml:"ldapSetting" xml:"ldapSetting"`
}

type xmlLdapConfig struct {
	XMLName      xml.Name     `xml:"config"`
	LdapSettings LdapSettings `xml:"security>ldapSettings"`
}

func resourceArtifactoryLdapSetting() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceLdapSettingsUpdate,
		CreateContext: resourceLdapSettingsUpdate,
		DeleteContext: resourceLdapSettingsDelete,
		ReadContext:   resourceLdapSettingsRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Ldap setting name",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Flag to enable or disable the ldap setting",
			},
			"ldap_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Location of the LDAP server in the following format: ldap://myserver:myport/dc=sampledomain,dc=com",
			},
			"user_dn_pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "A DN pattern that can be used to log users directly in to LDAP. This pattern is used to create a DN string for 'direct' user authentication where the pattern is relative to the base DN in the LDAP URL. The pattern argument {0} is replaced with the username. This only works if anonymous binding is allowed and a direct user DN can be used, which is not the default case for Active Directory (use User DN search filter instead). Example: uid={0},ou=People",
			},
			"auto_create_user": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When set, users are automatically created when using LDAP. Otherwise, users are transient and associated with auto-join groups defined in Artifactory.",
			},
			"email_attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIsEmail,
				Description:  "An attribute that can be used to map a user's email address to a user created automatically in Artifactory.",
			},
			"ldap_poisoning_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Protects against LDAP poisoning by filtering out users exposed to vulnerabilities.",
			},
			"allow_user_to_access_profile": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Auto created users will have access to their profile page and will be able to perform actions such as generating an API key.",
			},
			"paging_support_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When set, supports paging results for the LDAP server. This feature requires that the LDAP server supports a PagedResultsControl configuration.",
			},
			"search_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A filter expression used to search for the user DN used in LDAP authentication. This is an LDAP search filter (as defined in 'RFC 2254') with optional arguments. In this case, the username is the only argument, and is denoted by '{0}'. Possible examples are: (uid={0}) - This searches for a username match on the attribute. Authentication to LDAP is performed from the DN found if successful.",
			},
			"search_base": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "(Optional) A context name to search in relative to the base DN of the LDAP URL. For example, 'ou=users' With the LDAP Group Add-on enabled, it is possible to enter multiple search base entries separated by a pipe ('|') character.",
			},
			"search_sub_tree": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When set, enables deep search through the sub tree of the LDAP URL + search base.",
			},
			"manager_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The full DN of the user that binds to the LDAP server to perform user searches. Only used with \"search\" authentication.",
			},
			"manager_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The password of the user that binds to the LDAP server to perform the search. Only used with \"search\" authentication.",
				Sensitive:   true,
				Computed:    true,
			},
		},
	}
}

func resourceLdapSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rClient := m.(*resty.Client)
	ldapConfigs := &xmlLdapConfig{}
	rsrcLdapsetting := unpackLdapSetting(d)

	resp, err := rClient.R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/system/configuration during Read")
	}

	err = xml.Unmarshal(resp.Body(), &ldapConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap setting during read operation")
	}
	matchedLdapSetting := LdapSetting{}
	for _, iterLdapSetting := range ldapConfigs.LdapSettings.LdapSettingArr {
		if iterLdapSetting.Key == rsrcLdapsetting.Key {
			matchedLdapSetting = iterLdapSetting
		}
	}

	packDiag := packLdapSetting(&matchedLdapSetting, d)
	if packDiag != nil {
		return packDiag
	}
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Usage of Undocumented Artifactory API Endpoints",
		Detail:   "The artifactory_ldap_setting resource uses endpoints that are undocumented and may not work with SaaS environments, or may change without notice.",
	}}
}

func resourceLdapSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpackedLdapSetting := unpackLdapSetting(d)

	var constructBody = map[string]map[string]map[string]LdapSetting{}
	constructBody["security"] = map[string]map[string]LdapSetting{}
	constructBody["security"]["ldapSettings"] = map[string]LdapSetting{}
	constructBody["security"]["ldapSettings"][unpackedLdapSetting.Key] = *unpackedLdapSetting
	content, err := yaml.Marshal(&constructBody)

	if err != nil {
		return diag.Errorf("failed to marshal ldap settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one ldap setting resource, using same id
	d.SetId(unpackedLdapSetting.Key)
	return resourceLdapSettingsRead(ctx, d, m)
}

func resourceLdapSettingsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rClient := m.(*resty.Client)
	ldapConfigs := &xmlLdapConfig{}

	rsrcLdapSetting := unpackLdapSetting(d)

	response, err := rClient.R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from <base_url>/artifactory/api/system/configuration during Read")
	}

	err = xml.Unmarshal(response.Body(), &ldapConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap setting during delete operation")
	}

	var restoreLdapSettings = map[string]map[string]map[string]LdapSetting{}
	restoreLdapSettings["security"] = map[string]map[string]LdapSetting{}
	restoreLdapSettings["security"]["ldapSettings"] = map[string]LdapSetting{}

	for _, iterLdapSetting := range ldapConfigs.LdapSettings.LdapSettingArr {
		if iterLdapSetting.Key != rsrcLdapSetting.Key {
			restoreLdapSettings["security"]["ldapSettings"][iterLdapSetting.Key] = iterLdapSetting
		}
	}

	var clearAllLdapSettingsConfigs = `
security:
  ldapSettings: ~
`
	err = sendConfigurationPatch([]byte(clearAllLdapSettingsConfigs), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete for clearing all Ldap Settings")
	}

	restoreRestOfLdapSettingsConfigs, err := yaml.Marshal(&restoreLdapSettings)
	if err != nil {
		return diag.Errorf("failed to marshal ldap settings during Update")
	}

	err = sendConfigurationPatch([]byte(restoreRestOfLdapSettingsConfigs), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during restoration of Ldap Settings")
	}
	return nil
}

func unpackLdapSetting(s *schema.ResourceData) *LdapSetting {
	d := &ResourceData{s}
	ldapSetting := new(LdapSetting)
	ldapSetting.Key = d.getString("key", false)
	ldapSetting.Enabled = d.getBool("enabled", false)
	ldapSetting.LdapUrl = d.getString("ldap_url", false)
	ldapSetting.AutoCreateUser = d.getBool("auto_create_user", false)
	ldapSetting.LdapPoisoningProtection = d.getBool("ldap_poisoning_protection", false)
	ldapSetting.PagingSupportEnabled = d.getBool("paging_support_enabled", false)
	ldapSetting.AllowUserToAccessProfile = d.getBool("allow_user_to_access_profile", false)
	ldapSetting.UserDnPattern = d.getString("user_dn_pattern", false)
	ldapSetting.EmailAttribute = d.getString("email_attribute", false)
	ldapSetting.Search.SearchSubTree = d.getBool("search_sub_tree", false)
	ldapSetting.Search.SearchBase = d.getString("search_base", false)
	ldapSetting.Search.SearchFilter = d.getString("search_filter", false)
	ldapSetting.Search.ManagerDn = d.getString("manager_dn", false)
	ldapSetting.Search.ManagerPassword = d.getString("manager_password", true)
	return ldapSetting
}

func packLdapSetting(ldapSetting *LdapSetting, d *schema.ResourceData) diag.Diagnostics {
	setValue := mkLens(d)
	setValue("key", ldapSetting.Key)
	setValue("ldap_url", ldapSetting.LdapUrl)
	setValue("enabled", ldapSetting.Enabled)
	setValue("user_dn_pattern", ldapSetting.UserDnPattern)
	setValue("auto_create_user", ldapSetting.AutoCreateUser)
	setValue("ldap_poisoning_protection", ldapSetting.LdapPoisoningProtection)
	setValue("allow_user_to_access_profile", ldapSetting.AllowUserToAccessProfile)
	setValue("paging_support_enabled", ldapSetting.PagingSupportEnabled)
	setValue("search_base", ldapSetting.Search.SearchBase)
	setValue("search_filter", ldapSetting.Search.SearchFilter)
	setValue("search_sub_tree", ldapSetting.Search.SearchSubTree)
	setValue("manager_dn", ldapSetting.Search.ManagerDn)
	errors := setValue("email_attribute", ldapSetting.EmailAttribute)
	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack ldap settings %q", errors)
	}
	return nil
}
