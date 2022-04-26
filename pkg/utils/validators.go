package utils

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"gopkg.in/ldap.v2"
)

func RepoLayoutRefSchemaOverrideValidator(_ interface{}, _ cty.Path) diag.Diagnostics {
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Always override repo_layout_ref attribute in the schema",
			Detail:   "Always override repo_layout_ref attribute in the schema on top of base schema",
		},
	}
}

func ValidateLdapDn(value interface{}, _ string) ([]string, []error) {
	_, err := ldap.ParseDN(value.(string))
	if err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

func ValidateLdapFilter(value interface{}, _ string) ([]string, []error) {
	_, err := ldap.CompileFilter(value.(string))
	if err != nil {
		return nil, []error{err}
	}
	return nil, nil
}
