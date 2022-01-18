package artifactory

import (
	"context"
	"encoding/xml"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

type XmlLdapConfig struct {
	XMLName      xml.Name     `xml:"config"`
	LdapSettings LdapSettings `xml:"security>ldapSettings"`
}

var ldap_setting_schema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) Ldap setting name.`,
	},
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `(Optional) Flag to enable or disable the ldap setting. Default value is "true".`,
	},
	"ldap_url": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsURLWithScheme([]string{"ldap", "ldaps"}),
		Description:  "(Required) Location of the LDAP server in the following format: ldap://myldapserver/dc=sampledomain,dc=com",
	},
	"user_dn_pattern": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  "(Required) A DN pattern that can be used to log users directly in to LDAP. This pattern is used to create a DN string for 'direct' user authentication where the pattern is relative to the base DN in the LDAP URL. The pattern argument {0} is replaced with the username. This only works if anonymous binding is allowed and a direct user DN can be used, which is not the default case for Active Directory (use User DN search filter instead). Example: uid={0},ou=People",
	},
	"auto_create_user": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `(Optional) When set, users are automatically created when using LDAP. Otherwise, users are transient and associated with auto-join groups defined in Artifactory. Default value is "true".`,
	},
	"email_attribute": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validateIsEmail,
		Description:  `(Optional) An attribute that can be used to map a user's email address to a user created automatically in Artifactory.`,
	},
	"ldap_poisoning_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `(Optional) Protects against LDAP poisoning by filtering out users exposed to vulnerabilities. Default value is "true".`,
	},
	"allow_user_to_access_profile": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `(Optional) Auto created users will have access to their profile page and will be able to perform actions such as generating an API key. Default value is "false".`,
	},
	"paging_support_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `(Optional) When set, supports paging results for the LDAP server. This feature requires that the LDAP server supports a PagedResultsControl configuration. Default value is "true".`,
	},
	"search_filter": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "(Optional) A filter expression used to search for the user DN used in LDAP authentication. This is an LDAP search filter (as defined in 'RFC 2254') with optional arguments. In this case, the username is the only argument, and is denoted by '{0}'. Possible examples are: (uid={0}) - This searches for a username match on the attribute. Authentication to LDAP is performed from the DN found if successful.",
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
		Description: `(Optional) When set, enables deep search through the sub tree of the LDAP URL + search base. Default value is "true".`,
	},
	"manager_dn": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: `(Optional) The full DN of the user that binds to the LDAP server to perform user searches. Only used with "search" authentication.`,
	},
	"manager_password": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: `(Optional) The password of the user that binds to the LDAP server to perform the search. Only used with "search" authentication.`,
		Sensitive:   true,
		Computed:    true,
	},
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

		Schema: ldap_setting_schema,
	}
}

func resourceLdapSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ldapConfigs := &XmlLdapConfig{}
	ldapSetting := unpackLdapSetting(d)

	resp, err := m.(*resty.Client).R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
	}

	err = xml.Unmarshal(resp.Body(), &ldapConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap setting during read operation")
	}
	matchedLdapSetting := LdapSetting{}
	for _, iterLdapSetting := range ldapConfigs.LdapSettings.LdapSettingArr {
		if iterLdapSetting.Key == ldapSetting.Key {
			matchedLdapSetting = iterLdapSetting
			break
		}
	}
	var ldapSettingClass = ignoreHclPredicate("class", "rclass", "manager_password")
	packer := universalPack(
		allHclPredicate(
			ldapSettingClass, schemaHasKey(ldap_setting_schema),
		),
	)
	return diag.FromErr(packer(&matchedLdapSetting, d))
}

func resourceLdapSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpackedLdapSetting := unpackLdapSetting(d)

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	GET call structure has "security -> ldapSettings -> ldapSetting -> Array of ldapSetting config blocks".
	PATCH call structure has "security -> ldapSettings -> Name/Key of ldap setting that is being patch -> config block of the ldapSetting being patched".
	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.
	*/
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
	ldapConfigs := &XmlLdapConfig{}

	rsrcLdapSetting := unpackLdapSetting(d)

	response, err := m.(*resty.Client).R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
	}
	if response.IsError() {
		return diag.Errorf("Got error response for API: /artifactory/api/system/configuration request during Read")
	}

	err = xml.Unmarshal(response.Body(), &ldapConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap setting during delete operation")
	}

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	GET call structure has "security -> ldapSettings -> ldapSetting -> Array of ldapSetting config blocks".
	PATCH call structure has "security -> ldapSettings -> Name/Key of ldap setting that is being patch -> config block of the ldapSetting being patched".
	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.
	*/
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
	ldapSetting := *new(LdapSetting)
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
	return &ldapSetting
}
