package security

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/retry"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/validators"
	"net/http"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func ResourceArtifactoryUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Exists:        resourceUserExists,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidateIsEmail,
			},
			"admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"profile_updatable": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"disable_ui_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"internal_password_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
				ValidateFunc: func(tfValue interface{}, key string) ([]string, []error) {
					validationOn, _ := strconv.ParseBool(os.Getenv("JFROG_PASSWD_VALIDATION_ON"))
					if validationOn {
						ses, err := validators.DefaultPassValidation(tfValue, key)
						if err != nil {
							return append(ses, "if your organization has custom password rules, you may override "+
								"password validators by setting env var JFROG_PASSWD_VALIDATION_ON=false"), append(err)
						}
					}
					return nil, nil
				},
				StateFunc: func(str interface{}) string {
					// Avoid storing the actual value in the state and instead store the hash of it
					value, ok := str.(string)
					if !ok {
						panic(fmt.Errorf("'str' is not a string %s", str))
					}
					hash := sha256.Sum256([]byte(value))
					return base64.StdEncoding.EncodeToString(hash[:])
				},
			},
		},
	}
}

func resourceUserExists(data *schema.ResourceData, m interface{}) (bool, error) {

	d := &util.ResourceData{data}
	name := d.Id()
	return userExists(m.(*resty.Client), name)
}

func userExists(client *resty.Client, userName string) (bool, error) {
	_, err := client.R().Head("artifactory/api/security/users/" + userName)
	return err == nil, err
}

func unpackUser(s *schema.ResourceData) services.User {
	d := &util.ResourceData{s}
	return services.User{
		Name:                     d.GetString("name", false),
		Email:                    d.GetString("email", false),
		Password:                 d.GetString("password", true),
		Admin:                    d.GetBool("admin", false),
		ProfileUpdatable:         d.GetBool("profile_updatable", false),
		DisableUIAccess:          d.GetBool("disable_ui_access", false),
		InternalPasswordDisabled: d.GetBool("internal_password_disabled", false),
		Groups:                   d.GetSet("groups"),
	}
}

func packUser(user services.User, d *schema.ResourceData) error {

	setValue := util.MkLens(d)

	setValue("name", user.Name)
	setValue("email", user.Email)
	setValue("admin", user.Admin)
	setValue("profile_updatable", user.ProfileUpdatable)
	setValue("disable_ui_access", user.DisableUIAccess)
	errors := setValue("internal_password_disabled", user.InternalPasswordDisabled)

	if user.Groups != nil {
		errors = setValue("groups", schema.NewSet(schema.HashString, util.CastToInterfaceArr(user.Groups)))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack user %q", errors)
	}

	return nil
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	user := unpackUser(d)

	if user.Name == "" {
		return diag.Errorf("user name cannot be empty")
	}

	if user.Password == "" {
		return diag.Errorf("no password supplied. Please use any of the terraform random password generators")
	}
	_, err := m.(*resty.Client).R().SetBody(user).AddRetryCondition(retry.On404NotFound).Put("artifactory/api/security/users/" + user.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Name)
	return nil
}

func resourceUserRead(ctx context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	d := &util.ResourceData{ResourceData: rd}

	userName := d.Id()
	user := &services.User{}
	resp, err := m.(*resty.Client).R().SetResult(user).Get("artifactory/api/security/users/" + userName)

	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(packUser(*user, rd))
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	user := unpackUser(d)
	_, err := m.(*resty.Client).R().SetBody(user).Post("artifactory/api/security/users/" + user.Name)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Name)
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(_ context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	d := &util.ResourceData{ResourceData: rd}
	userName := d.GetString("name", false)

	_, err := m.(*resty.Client).R().Delete("artifactory/api/security/users/" + userName)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
