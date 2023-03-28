package configuration

import (
	"context"
	"encoding/xml"

	"github.com/jfrog/terraform-provider-shared/packer"

	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

type LdapGroupSetting struct {
	Name                 string `xml:"name" yaml:"name"`
	EnabledLdap          string `hcl:"ldap_setting_key" xml:"enabledLdap" yaml:"enabledLdap"`
	GroupBaseDn          string `xml:"groupBaseDn" yaml:"groupBaseDn"`
	GroupNameAttribute   string `xml:"groupNameAttribute" yaml:"groupNameAttribute"`
	GroupMemberAttribute string `xml:"groupMemberAttribute" yaml:"groupMemberAttribute"`
	SubTree              bool   `xml:"subTree" yaml:"subTree"`
	Filter               string `xml:"filter" yaml:"filter"`
	DescriptionAttribute string `xml:"descriptionAttribute" yaml:"descriptionAttribute"`
	Strategy             string `xml:"strategy" yaml:"strategy"`
}

func (l LdapGroupSetting) Id() string {
	return l.Name
}

type LdapGroupSettings struct {
	LdapGroupSettingArr []LdapGroupSetting `xml:"ldapGroupSetting" yaml:"ldapGroupSetting"`
}

type SecurityLdapGroupSettings struct {
	LdapGroupSettings LdapGroupSettings `xml:"ldapGroupSettings"`
}

type XmlLdapGroupConfig struct {
	XMLName  xml.Name                  `xml:"config"`
	Security SecurityLdapGroupSettings `xml:"security"`
}

func ResourceArtifactoryLdapGroupSetting() *schema.Resource {
	var ldapGroupSettingsSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `Ldap group setting name.`,
		},
		"ldap_setting_key": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `The LDAP setting key you want to use for group retrieval. The value for this field corresponds to 'enabledLdap' field of the ldap group setting XML block of system configuration.`,
		},
		"group_base_dn": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "",
			ValidateDiagFunc: validator.LdapDn,
			Description:      `A search base for group entry DNs, relative to the DN on the LDAP server’s URL (and not relative to the LDAP Setting’s “Search Base”). Used when importing groups.`,
		},
		"group_name_attribute": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Attribute on the group entry denoting the group name. Used when importing groups.",
		},
		"group_member_attribute": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `A multi-value attribute on the group entry containing user DNs or IDs of the group members (e.g., uniqueMember,member).`,
		},
		"sub_tree": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `When set, enables deep search through the sub-tree of the LDAP URL + Search Base. True by default.`,
		},
		"filter": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validator.All(validator.StringIsNotEmpty, validator.LdapFilter),
			Description:      `The LDAP filter used to search for group entries. Used for importing groups.`,
		},
		"description_attribute": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `An attribute on the group entry which denoting the group description. Used when importing groups.`,
		},
		"strategy": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"STATIC", "DYNAMIC", "HIERARCHICAL"}, false)),
			Description: `The JFrog Platform Deployment (JPD) supports three ways of mapping groups to LDAP schemas:
Static: Group objects are aware of their members, however, the users are not aware of the groups they belong to. Each group object such as groupOfNames or groupOfUniqueNames holds its respective member attributes, typically member or uniqueMember, which is a user DN.
Dynamic: User objects are aware of what groups they belong to, but the group objects are not aware of their members. Each user object contains a custom attribute, such as group, that holds the group DNs or group names of which the user is a member.
Hierarchy: The user's DN is indicative of the groups the user belongs to by using group names as part of user DN hierarchy. Each user DN contains a list of ou's or custom attributes that make up the group association. For example, uid=user1,ou=developers,ou=uk,dc=jfrog,dc=org indicates that user1 belongs to two groups: uk and developers.`,
		},
	}

	var resourceLdapGroupSettingsRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		data := &util.ResourceData{ResourceData: d}
		name := data.GetString("name", false)

		ldapGroupConfigs := XmlLdapGroupConfig{}
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&ldapGroupConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedLdapGroupSetting := FindConfigurationById[LdapGroupSetting](ldapGroupConfigs.Security.LdapGroupSettings.LdapGroupSettingArr, name)
		if matchedLdapGroupSetting == nil {
			d.SetId("")
			return nil
		}

		pkr := packer.Default(ldapGroupSettingsSchema)

		return diag.FromErr(pkr(matchedLdapGroupSetting, d))
	}

	var resourceLdapGroupSettingsUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		constructBody["security"]["ldapGroupSettings"][unpackedLdapGroupSetting.Name] = unpackedLdapGroupSetting
		content, err := yaml.Marshal(&constructBody)

		if err != nil {
			return diag.Errorf("failed to marshal ldap group settings during Update")
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Update")
		}

		// we should only have one ldap group setting resource, using same id
		d.SetId(unpackedLdapGroupSetting.Name)
		return resourceLdapGroupSettingsRead(ctx, d, m)
	}

	var resourceLdapGroupSettingsDelete = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		ldapGroupConfigs := &XmlLdapGroupConfig{}

		rsrcLdapGroupSetting := unpackLdapGroupSetting(d)

		response, err := m.(util.ProvderMetadata).Client.R().SetResult(&ldapGroupConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return diag.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
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

		for _, iterLdapGroupSetting := range ldapGroupConfigs.Security.LdapGroupSettings.LdapGroupSettingArr {
			if iterLdapGroupSetting.Name != rsrcLdapGroupSetting.Name {
				restoreLdapGroupSettings["security"]["ldapGroupSettings"][iterLdapGroupSetting.Name] = iterLdapGroupSetting
			}
		}

		var clearAllLdapGroupSettingsConfigs = `
security:
  ldapGroupSettings: ~
`
		err = SendConfigurationPatch([]byte(clearAllLdapGroupSettingsConfigs), m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Delete for clearing all Ldap Group Settings")
		}

		restoreRestOfLdapGroupSettingsConfigs, err := yaml.Marshal(&restoreLdapGroupSettings)
		if err != nil {
			return diag.Errorf("failed to marshal ldap group settings during Update")
		}

		err = SendConfigurationPatch(restoreRestOfLdapGroupSettingsConfigs, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during restoration of Ldap Group Settings")
		}
		return nil
	}

	return &schema.Resource{
		UpdateContext: resourceLdapGroupSettingsUpdate,
		CreateContext: resourceLdapGroupSettingsUpdate,
		DeleteContext: resourceLdapGroupSettingsDelete,
		ReadContext:   resourceLdapGroupSettingsRead,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema:      ldapGroupSettingsSchema,
		Description: "Provides an Artifactory ldap group setting resource. This resource configuration corresponds to ldapGroupSettings config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}

func unpackLdapGroupSetting(s *schema.ResourceData) LdapGroupSetting {
	d := &util.ResourceData{ResourceData: s}
	ldapGroupSetting := LdapGroupSetting{
		Name:                 d.GetString("name", false),
		EnabledLdap:          d.GetString("ldap_setting_key", false),
		GroupBaseDn:          d.GetString("group_base_dn", false),
		GroupNameAttribute:   d.GetString("group_name_attribute", false),
		GroupMemberAttribute: d.GetString("group_member_attribute", false),
		SubTree:              d.GetBool("sub_tree", false),
		Filter:               d.GetString("filter", false),
		DescriptionAttribute: d.GetString("description_attribute", false),
		Strategy:             d.GetString("strategy", false),
	}
	return ldapGroupSetting
}
