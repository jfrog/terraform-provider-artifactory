package artifactory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type User struct {
	Name                     string   `json:"name"`
	Email                    string   `json:"email"`
	Password                 string   `json:"password"`
	Admin                    bool     `json:"admin"`
	ProfileUpdatable         bool     `json:"profileUpdatable"`
	DisableUIAccess          bool     `json:"disableUIAccess"`
	InternalPasswordDisabled bool     `json:"internalPasswordDisabled"`
	LastLoggedIn             string   `json:"lastLoggedIn"`
	Realm                    string   `json:"realm"`
	Groups                   []string `json:"groups"`
}

func resourceArtifactoryUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Exists: resourceUserExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Username for user.",
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateIsEmail,
				Description:  "Email for user.",
			},
			"admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When enabled, this user is an administrator with all the ensuing privileges.",
			},
			"profile_updatable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "When enabled, this user can update their profile details (except for the password. " +
					"Only an administrator can update the password). There may be cases in which you want to leave " +
					"this unset to prevent users from updating their profile. For example, a departmental user with " +
					"a single password shared between all department members.",
			},
			"disable_ui_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "When enabled, this user can only access the system through the REST API." +
					" This option cannot be set if the user has Admin privileges.",
			},
			"internal_password_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "When enabled, disables the fallback mechanism for using an internal password when " +
					"external authentication (such as LDAP) is enabled.",
			},
			"groups": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of groups this user is a part of.",
			},
			"password": {
				Type:             schema.TypeString,
				Sensitive:        true,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description: "Password for the user. Password validation is not done by the provider and is " +
					"offloaded onto the Artifactory.",
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

	d := &ResourceData{data}
	name := d.Id()
	return userExists(m.(*resty.Client), name)
}

func userExists(client *resty.Client, userName string) (bool, error) {
	resp, err := client.R().Head("artifactory/api/security/users/" + userName)
	if err != nil && resp != nil && resp.StatusCode() == http.StatusNotFound {
		// Do not error on 404s as this causes errors when the upstream user has been manually removed
		return false, nil
	}

	return err == nil, err
}

func unpackUser(s *schema.ResourceData) User {
	d := &ResourceData{s}
	return User{
		Name:                     d.getString("name", false),
		Email:                    d.getString("email", false),
		Password:                 d.getString("password", true),
		Admin:                    d.getBool("admin", false),
		ProfileUpdatable:         d.getBool("profile_updatable", false),
		DisableUIAccess:          d.getBool("disable_ui_access", false),
		InternalPasswordDisabled: d.getBool("internal_password_disabled", false),
		Groups:                   d.getSet("groups"),
	}
}

func packUser(user User, d *schema.ResourceData) error {

	setValue := mkLens(d)

	setValue("name", user.Name)
	setValue("email", user.Email)
	setValue("admin", user.Admin)
	setValue("profile_updatable", user.ProfileUpdatable)
	setValue("disable_ui_access", user.DisableUIAccess)
	errors := setValue("internal_password_disabled", user.InternalPasswordDisabled)

	if user.Groups != nil {
		errors = setValue("groups", schema.NewSet(schema.HashString, castToInterfaceArr(user.Groups)))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack user %q", errors)
	}

	return nil
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	user := unpackUser(d)

	if user.Name == "" {
		return fmt.Errorf("user name cannot be empty")
	}

	if user.Password == "" {
		return fmt.Errorf("no password supplied. Please use any of the terraform random password generators")
	}
	_, err := m.(*resty.Client).R().SetBody(user).Put("artifactory/api/security/users/" + user.Name)
	if err != nil {
		return err
	}

	d.SetId(user.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result := &User{}
		resp, e := m.(*resty.Client).R().SetResult(result).Get("artifactory/api/security/users/" + user.Name)

		if e != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return resource.RetryableError(fmt.Errorf("expected user to be created, but currently not found"))
			}
			return resource.NonRetryableError(fmt.Errorf("error describing user: %s", err))
		}

		return nil
	})
}

func resourceUserRead(rd *schema.ResourceData, m interface{}) error {
	d := &ResourceData{rd}

	userName := d.Id()
	user := &User{}
	resp, err := m.(*resty.Client).R().SetResult(user).Get("artifactory/api/security/users/" + userName)

	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
	return packUser(*user, rd)
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	user := unpackUser(d)
	_, err := m.(*resty.Client).R().SetBody(user).Post("artifactory/api/security/users/" + user.Name)

	if err != nil {
		return err
	}

	d.SetId(user.Name)
	return resourceUserRead(d, m)
}

func resourceUserDelete(rd *schema.ResourceData, m interface{}) error {
	d := &ResourceData{rd}
	userName := d.getString("name", false)

	_, err := m.(*resty.Client).R().Delete("artifactory/api/security/users/" + userName)
	if err != nil {
		return fmt.Errorf("user %s not deleted. %s", userName, err)
	}
	return nil
}
