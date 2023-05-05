package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/exp/maps"
)

func ResourceArtifactoryUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
		"password": {
			Type:      schema.TypeString,
			Sensitive: true,
			Optional:  true,
			Description: "(Optional, Sensitive) Password for the user. When omitted, a random password is generated using the following password policy: " +
				"12 characters with 1 digit, 1 symbol, with upper and lower case letters",
		},
	}
	maps.Copy(userSchema, baseUserSchema)

	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: userSchema,

		Description: "Provides an Artifactory user resource. This can be used to create and manage Artifactory users. The password is a required field by the [Artifactory API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplaceUser), but we made it optional in this resource to accommodate the scenario where the password is not needed and will be reset by the actual user later. When the optional attribute `password` is omitted, a random password is generated according to current Artifactory password policy.",
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	passwordGenerator := func(user *User) diag.Diagnostics {
		var diags diag.Diagnostics

		if user.Password == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "No password supplied",
				Detail:   "One will be generated (12 characters with 1 digit, 1 symbol, with upper and lower case letters) and this may fail as your Artifactory password policy can't be determined by the provider.",
			})

			// Generate a password that is 12 characters long with 1 digit, 1 symbol,
			// allowing upper and lower case letters, disallowing repeat characters.
			randomPassword, err := password.Generate(12, 1, 1, false, false)
			if err != nil {
				return diag.Errorf("failed to generate password. %v", err)
			}

			user.Password = randomPassword
		}

		return diags
	}

	return resourceBaseUserCreate(ctx, d, m, passwordGenerator)
}
