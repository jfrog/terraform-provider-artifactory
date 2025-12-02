// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceArtifactorySamlSettings() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceSamlSettingsUpdate,
		CreateContext: resourceSamlSettingsUpdate,
		DeleteContext: resourceSamlSettingsDelete,
		ReadContext:   resourceSamlSettingsRead,

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `Enable SAML SSO.  Default value is "true".`,
			},
			"certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `SAML certificate that contains the public key for the IdP service provider.  Used by Artifactory to verify sign-in requests. Default value is "".`,
			},
			"email_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `Name of the attribute in the SAML response from the IdP that contains the user's email. Default value is "".`,
			},
			"group_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: `Name of the attribute in the SAML response from the IdP that contains the user's group memberships. Default value is "".`,
			},
			"login_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Service provider login url configured on the IdP.`,
			},
			"logout_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Service provider logout url, or where to redirect after user logs out.`,
			},
			"no_auto_user_creation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When automatic user creation is off, authenticated users are not automatically created inside Artifactory. Instead, for every request from an SSO user, the user is temporarily associated with default groups (if such groups are defined), and the permissions for these groups apply. Without auto-user creation, you must manually create the user inside Artifactory to manage user permissions not attached to their default groups. Default value is "false".`,
			},
			"service_provider_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `The SAML service provider name. This should be a URI that is also known as the entityID, providerID, or entity identity.`,
			},
			"allow_user_to_access_profile": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `Allow persisted users to access their profile.  Default value is "true".`,
			},
			"auto_redirect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Auto redirect to login through the IdP when clicking on Artifactory's login link.  Default value is "false".`,
			},
			"sync_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Associate user with Artifactory groups based on the "group_attribute" provided in the SAML response from the identity provider.  Default value is "false".`,
			},
			"verify_audience_restriction": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `Enable "audience", or who the SAML assertion is intended for.  Ensures that the correct service provider intended for Artifactory is used on the IdP. Default value is "true".`,
			},
			"use_encrypted_assertion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When set, an X.509 public certificate will be created by Artifactory. Download this certificate and upload it to your IDP and choose your own encryption algorithm. This process will let you encrypt the assertion section in your SAML response. Default value is "false".`,
			},
		},
		DeprecationMessage: `This resource is deprecated in favor of "platform_saml_settings" (https://registry.terraform.io/providers/jfrog/platform/latest/docs/resources/saml_settings) resource in the Platform provider.`,
	}
}

func resourceSamlSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_saml_settings deprecated. Use platform_saml_settings instead")
}

func resourceSamlSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_saml_settings deprecated. Use platform_saml_settings instead")
}

func resourceSamlSettingsDelete(_ context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_saml_settings deprecated. Use platform_saml_settings instead")
}
