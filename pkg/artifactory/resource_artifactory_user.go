package artifactory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/exp/maps"
)

func resourceArtifactoryUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Optional:    true,
			Description: "(Optional) Password for the user. When omitted, a random password is generated according to default Artifactory password policy.",
		},
	}
	maps.Copy(userSchema, baseUserSchema)

	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Exists:        resourceUserExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: userSchema,

		Description: "Provides an Artifactory unmanaged user resource. This can be used to create and manage Artifactory users. Password is optional and one will be automatically generated.",
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	passwordGenerator := func(user *User) diag.Diagnostics {
		var diags diag.Diagnostics

		if user.Password == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "No password supplied",
				Detail:   "One will be generated (10 characters with 1 digit, 1 symbol, with upper and lower case letters) and this may fail as your Artifactory password policy can't be determined by the provider.",
			})

			// Generate a password that is 10 characters long with 1 digit, 1 symbol,
			// allowing upper and lower case letters, disallowing repeat characters.
			randomPassword, err := password.Generate(10, 1, 1, false, false)
			if err != nil {
				return diag.Errorf("failed to generate password. %v", err)
			}

			user.Password = randomPassword
		}

		return diags
	}

	return resourceBaseUserCreate(ctx, d, m, passwordGenerator)
}
