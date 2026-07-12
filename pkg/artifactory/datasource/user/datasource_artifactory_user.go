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

package user

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func DataSourceArtifactoryUser() *schema.Resource {
	var userSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Username for user.",
		},
		"email": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validator.IsEmail,
			Description:      "Email for user.",
		},
		"admin": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "(Optional, Default: false) When enabled, this user is an administrator with all the ensuing privileges.",
		},
		"profile_updatable": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "(Optional, Default: true) When enabled, this user can update their profile details (except for the password. " +
				"Only an administrator can update the password). There may be cases in which you want to leave " +
				"this unset to prevent users from updating their profile. For example, a departmental user with " +
				"a single password shared between all department members.",
		},
		"disable_ui_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "(Optional, Default: true) When enabled, this user can only access the system through the REST API." +
				" This option cannot be set if the user has Admin privileges.",
		},
		"internal_password_disabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "(Optional, Default: false) When enabled, disables the fallback mechanism for using an internal password when " +
				"external authentication (such as LDAP) is enabled.",
		},
		"groups": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Set:         schema.HashString,
			Optional:    true,
			Description: "List of groups this user is a part of.",
		},
	}

	read := func(_ context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
		d := &utilsdk.ResourceData{ResourceData: rd}

		userName := d.Get("name").(string)
		var userObj User
		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := readUser(
			m.(util.ProviderMetadata).Client.R(),
			m.(util.ProviderMetadata).ArtifactoryVersion,
			userName,
			&userObj,
			&artifactoryError)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.IsError() {
			return diag.Errorf("%s", resp.String())
		}

		d.SetId(userObj.Name)

		return packUser(userObj, rd)
	}

	return &schema.Resource{
		ReadContext: read,
		Schema:      userSchema, // note this does not include password of the user, don't think we should return it as a datasource
		Description: "Provides the Artifactory User data source. ",
	}
}

type User struct {
	Name                     string   `json:"username"`
	Email                    string   `json:"email"`
	Password                 string   `json:"password,omitempty"`
	Admin                    bool     `json:"admin"`
	ProfileUpdatable         bool     `json:"profile_updatable"`
	DisableUIAccess          bool     `json:"disable_ui_access"`
	InternalPasswordDisabled *bool    `json:"internal_password_disabled"`
	Groups                   []string `json:"groups,omitempty"`
}

func readUser(req *resty.Request, artifactoryVersion, name string, result *User, artifactoryError *artifactory.ArtifactoryErrorsResponse) (*resty.Response, error) {
	endpoint := user.GetUserEndpointPath(artifactoryVersion)

	// 7.83.1 or later, use Access API
	if ok, err := util.CheckVersion(artifactoryVersion, user.AccessAPIArtifactoryVersion); err == nil && ok {
		return req.
			SetPathParam("name", name).
			SetResult(&result).
			SetError(&artifactoryError).
			Get(endpoint)
	}

	// else use old Artifactory API, which has a slightly differect JSON payload!
	var artifactoryResult user.ArtifactoryUserAPIModel
	res, err := req.
		SetPathParam("name", name).
		SetResult(&artifactoryResult).
		SetError(artifactoryError).
		Get(endpoint)

	var groups []string
	if artifactoryResult.Groups != nil {
		groups = *artifactoryResult.Groups
	}

	*result = User{
		Name:                     artifactoryResult.Name,
		Email:                    artifactoryResult.Email,
		Admin:                    artifactoryResult.Admin,
		ProfileUpdatable:         artifactoryResult.ProfileUpdatable,
		DisableUIAccess:          artifactoryResult.DisableUIAccess,
		InternalPasswordDisabled: artifactoryResult.InternalPasswordDisabled,
		Groups:                   groups,
	}

	return res, err
}

func packUser(user User, d *schema.ResourceData) diag.Diagnostics {

	setValue := utilsdk.MkLens(d)

	setValue("name", user.Name)
	setValue("email", user.Email)
	setValue("admin", user.Admin)
	setValue("profile_updatable", user.ProfileUpdatable)
	setValue("disable_ui_access", user.DisableUIAccess)
	errors := setValue("internal_password_disabled", user.InternalPasswordDisabled)

	if user.Groups != nil {
		errors = setValue("groups", schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(user.Groups)))
	}

	if len(errors) > 0 {
		return diag.Errorf("failed to pack user %q", errors)
	}

	return nil
}
