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

type LdapGroupSetting struct {
	Name                 string `xml:"name" yaml:"name"`
	EnabledLdap          string `xml:"enabledLdap" yaml:"enabledLdap"`
	GroupBaseDn          string `xml:"groupBaseDn" yaml:"groupBaseDn"`
	GroupNameAttribute   string `xml:"groupNameAttribute" yaml:"groupNameAttribute"`
	GroupMemberAttribute string `xml:"groupMemberAttribute" yaml:"groupMemberAttribute"`
	SubTree              bool   `xml:"subTree" yaml:"subTree"`
	Filter               string `xml:"filter" yaml:"filter"`
	DescriptionAttribute string `xml:"descriptionAttribute" yaml:"descriptionAttribute"`
	Strategy             string `xml:"strategy" yaml:"strategy"`
}

type LdapGroupSettings struct {
	LdapGroupSettingArr []LdapGroupSetting `yaml:"ldapGroupSetting" xml:"ldapGroupSetting"`
}

type XmlLdapGroupConfig struct {
	XMLName           xml.Name          `xml:"config"`
	LdapGroupSettings LdapGroupSettings `xml:"security>ldapGroupSettings"`
}

var ldap_group_setting_schema = map[string]*schema.Schema{
	"name": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) Ldap group setting name.`,
	},
	"enabled_ldap": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) The LDAP setting you want to use for group retrieval.`,
	},
	"group_base_dn": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Optional) A search base for group entry DNs, relative to the DN on the LDAP server’s URL (and not relative to the LDAP Setting’s “Search Base”). Used when importing groups.`,
	},
	"group_name_attribute": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  "(Required) Attribute on the group entry denoting the group name. Used when importing groups.",
	},
	"group_member_attribute": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) A multi-value attribute on the group entry containing user DNs or IDs of the group members (e.g., uniqueMember,member).`,
	},
	"sub_tree": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `(Optional) When set, enables deep search through the sub-tree of the LDAP URL + Search Base. True by default.`,
	},
	"filter": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) The LDAP filter used to search for group entries. Used for importing groups.`,
	},
	"description_attribute": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  `(Required) An attribute on the group entry which denoting the group description. Used when importing groups.`,
	},
	"strategy": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"STATIC", "DYNAMIC", "HIERARCHICAL"}, false)),
		Description: `(Required) The JFrog Platform Deployment (JPD) supports three ways of mapping groups to LDAP schemas:
Static: Group objects are aware of their members, however, the users are not aware of the groups they belong to. Each group object such as groupOfNames or groupOfUniqueNames holds its respective member attributes, typically member or uniqueMember, which is a user DN.
Dynamic: User objects are aware of what groups they belong to, but the group objects are not aware of their members. Each user object contains a custom attribute, such as group, that holds the group DNs or group names of which the user is a member.
Hierarchy: The user's DN is indicative of the groups the user belongs to by using group names as part of user DN hierarchy. Each user DN contains a list of ou's or custom attributes that make up the group association. For example, uid=user1,ou=developers,ou=uk,dc=jfrog,dc=org indicates that user1 belongs to two groups: uk and developers.`,
	},
}

func resourceArtifactoryLdapGroupSetting() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceLdapGroupSettingsUpdate,
		CreateContext: resourceLdapGroupSettingsUpdate,
		DeleteContext: resourceLdapGroupSettingsDelete,
		ReadContext:   resourceLdapGroupSettingsRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: ldap_group_setting_schema,
	}
}

func resourceLdapGroupSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ldapGroupConfigs := &XmlLdapGroupConfig{}
	ldapGroupSetting := unpackLdapGroupSetting(d)

	resp, err := m.(*resty.Client).R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
	}

	err = xml.Unmarshal(resp.Body(), &ldapGroupConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap group setting during read operation")
	}
	matchedLdapGroupSetting := LdapGroupSetting{}
	for _, iterLdapGroupSetting := range ldapGroupConfigs.LdapGroupSettings.LdapGroupSettingArr {
		if iterLdapGroupSetting.Name == ldapGroupSetting.Name {
			matchedLdapGroupSetting = iterLdapGroupSetting
			break
		}
	}
	packer := universalPack(
		allHclPredicate(
			noClass, schemaHasKey(ldap_group_setting_schema),
		),
	)
	return diag.FromErr(packer(&matchedLdapGroupSetting, d))
}

func resourceLdapGroupSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unpackedLdapGroupSetting := unpackLdapGroupSetting(d)

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	GET call structure has "security -> ldapGroupSettings -> ldapGroupSetting -> Array of ldapGroupSetting config blocks".
	PATCH call structure has "security -> ldapGroupSettings -> Name/Key of ldap group setting that is being patch -> config block of the ldapGroupSetting being patched".
	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.
	*/
	var constructBody = map[string]map[string]map[string]LdapGroupSetting{}
	constructBody["security"] = map[string]map[string]LdapGroupSetting{}
	constructBody["security"]["ldapGroupSettings"] = map[string]LdapGroupSetting{}
	constructBody["security"]["ldapGroupSettings"][unpackedLdapGroupSetting.Name] = *unpackedLdapGroupSetting
	content, err := yaml.Marshal(&constructBody)

	if err != nil {
		return diag.Errorf("failed to marshal ldap group settings during Update")
	}

	err = sendConfigurationPatch(content, m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Update")
	}

	// we should only have one ldap group setting resource, using same id
	d.SetId(unpackedLdapGroupSetting.Name)
	return resourceLdapGroupSettingsRead(ctx, d, m)
}

func resourceLdapGroupSettingsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ldapGroupConfigs := &XmlLdapGroupConfig{}

	rsrcLdapGroupSetting := unpackLdapGroupSetting(d)

	response, err := m.(*resty.Client).R().Get("artifactory/api/system/configuration")
	if err != nil {
		return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
	}
	if response.IsError() {
		return diag.Errorf("Got error response for API: /artifactory/api/system/configuration request during Read")
	}

	err = xml.Unmarshal(response.Body(), &ldapGroupConfigs)
	if err != nil {
		return diag.Errorf("failed to xml unmarshal ldap group setting during delete operation")
	}

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	GET call structure has "security -> ldapGroupSettings -> ldapGroupSetting -> Array of ldapGroupSetting config blocks".
	PATCH call structure has "security -> ldapGroupSettings -> Name/Key of ldap group setting that is being patch -> config block of the ldapGroupSetting being patched".
	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.
	*/
	var restoreLdapGroupSettings = map[string]map[string]map[string]LdapGroupSetting{}
	restoreLdapGroupSettings["security"] = map[string]map[string]LdapGroupSetting{}
	restoreLdapGroupSettings["security"]["ldapGroupSettings"] = map[string]LdapGroupSetting{}

	for _, iterLdapGroupSetting := range ldapGroupConfigs.LdapGroupSettings.LdapGroupSettingArr {
		if iterLdapGroupSetting.Name != rsrcLdapGroupSetting.Name {
			restoreLdapGroupSettings["security"]["ldapGroupSettings"][iterLdapGroupSetting.Name] = iterLdapGroupSetting
		}
	}

	var clearAllLdapGroupSettingsConfigs = `
security:
  ldapGroupSettings: ~
`
	err = sendConfigurationPatch([]byte(clearAllLdapGroupSettingsConfigs), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during Delete for clearing all Ldap Group Settings")
	}

	restoreRestOfLdapGroupSettingsConfigs, err := yaml.Marshal(&restoreLdapGroupSettings)
	if err != nil {
		return diag.Errorf("failed to marshal ldap group settings during Update")
	}

	err = sendConfigurationPatch([]byte(restoreRestOfLdapGroupSettingsConfigs), m)
	if err != nil {
		return diag.Errorf("failed to send PATCH request to Artifactory during restoration of Ldap Group Settings")
	}
	return nil
}

func unpackLdapGroupSetting(s *schema.ResourceData) *LdapGroupSetting {
	d := &ResourceData{s}
	ldapGroupSetting := *new(LdapGroupSetting)
	ldapGroupSetting.Name = d.getString("name", false)
	ldapGroupSetting.EnabledLdap = d.getString("enabled_ldap", false)
	ldapGroupSetting.GroupBaseDn = d.getString("group_base_dn", false)
	ldapGroupSetting.GroupNameAttribute = d.getString("group_name_attribute", false)
	ldapGroupSetting.GroupMemberAttribute = d.getString("group_member_attribute", false)
	ldapGroupSetting.SubTree = d.getBool("sub_tree", false)
	ldapGroupSetting.Filter = d.getString("filter", false)
	ldapGroupSetting.DescriptionAttribute = d.getString("description_attribute", false)
	ldapGroupSetting.Strategy = d.getString("strategy", false)
	return &ldapGroupSetting
}
