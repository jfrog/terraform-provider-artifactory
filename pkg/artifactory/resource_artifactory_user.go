package artifactory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var warning = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

func resourceArtifactoryUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Exists: userExists,

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
						validate := composeValidators(containsDigit, containsLower, containsUpper, minLength)
						ses, err := validate(tfValue, key)
						if err != nil {
							return ses, append(err,
								fmt.Errorf("if your organization has custom password rules, you may override "+
									"password validation by setting env var JFROG_PASSWD_VALIDATION_ON=false"),
							)
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

func userExists(data *schema.ResourceData, config interface{}) (bool, error) {
	client := config.(*ArtClient).Resty
	d := &ResourceData{data}
	userName := d.getString("name", false)
	if userName == "" {
		return false, fmt.Errorf("no usersname supplied")
	}
	_, err := client.R().Head("artifactory/api/security/users/" + userName)
	if err != nil {
		return false, err
	}
	return true, nil
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
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("name", user.Name))
	logErr(d.Set("email", user.Email))
	logErr(d.Set("admin", user.Admin))
	logErr(d.Set("profile_updatable", user.ProfileUpdatable))
	logErr(d.Set("disable_ui_access", user.DisableUIAccess))
	logErr(d.Set("internal_password_disabled", user.InternalPasswordDisabled))

	if user.Groups != nil {
		logErr(d.Set("groups", schema.NewSet(schema.HashString, castToInterfaceArr(user.Groups))))
	}

	if hasErr {
		return fmt.Errorf("failed to pack user")
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

		return nil
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
