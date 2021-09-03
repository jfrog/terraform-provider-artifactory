package artifactory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"net/http"
	"os"
	"strconv"
)

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
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateIsEmail,
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
				Required:  true,
				ValidateFunc: func(tfValue interface{}, key string) ([]string, []error) {
					validationOn, _ := strconv.ParseBool(os.Getenv("JFROG_PASSWD_VALIDATION_ON"))
					if validationOn {
						validate := validation.All(containsDigit, containsLower, containsUpper, minLength)
						ses, err := validate(tfValue, key)
						if err != nil {
							return append(ses, "if your organization has custom password rules, you may override "+
								"password validation by setting env var JFROG_PASSWD_VALIDATION_ON=false"), append(err)
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

	d := &ResourceData{data}
	userName := d.getString("name", false)
	if userName == "" {
		return false, fmt.Errorf("no usersname supplied")
	}
	return userExists(m.(*ArtClient).Resty, userName)
}

func userExists(client *resty.Client, userName string) (bool, error) {
	_, err := client.R().Head("artifactory/api/security/users/" + userName)
	return err == nil, err
}

func unpackUser(s *schema.ResourceData) services.User {
	d := &ResourceData{s}
	return services.User{
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

func packUser(user services.User, d *schema.ResourceData) error {

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
	client := m.(*ArtClient).Resty

	user := unpackUser(d)

	if user.Name == "" {
		return fmt.Errorf("user name cannot be empty")
	}

	if user.Password == "" {
		return fmt.Errorf("no password supplied. Please use any of the terraform random password generators")
	}
	_, err := client.R().SetBody(user).Put("artifactory/api/security/users/" + user.Name)
	if err != nil {
		return err
	}

	d.SetId(user.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result := &services.User{}
		resp, e := client.R().SetResult(result).Get("artifactory/api/security/users/" + user.Name)

		if e != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return resource.RetryableError(fmt.Errorf("expected user to be created, but currently not found"))
			}
			return resource.NonRetryableError(fmt.Errorf("error describing user: %s", err))
		}
		return resource.NonRetryableError(resourceUserRead(d, m))
	})
}

func resourceUserRead(rd *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	d := &ResourceData{rd}

	userName := d.getString("name", false)
	user := &services.User{}
	resp, err := client.R().SetResult(user).Get("artifactory/api/security/users/" + userName)

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
	client := m.(*ArtClient).Resty

	user := unpackUser(d)
	_, err := client.R().SetBody(user).Post("artifactory/api/security/users/" + user.Name)

	if err != nil {
		return err
	}

	d.SetId(user.Name)
	return resourceUserRead(d, m)
}

func resourceUserDelete(rd *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	d := &ResourceData{rd}
	userName := d.getString("name", false)

	_, err := client.R().Delete("artifactory/api/security/users/" + userName)
	if err != nil {
		return fmt.Errorf("user %s not deleted. %s", userName, err)
	}
	return nil
}
