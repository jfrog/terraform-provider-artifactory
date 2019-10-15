package artifactory

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"math/rand"
	"net/http"
)

const randomPasswordLength = 16

func resourceArtifactoryUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
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
				StateFunc: func(value interface{}) string {
					// Avoid storing the actual value in the state and instead store the hash of it
					return hashString(value.(string))
				},
			},
		},
	}
}

// hashString do a sha256 checksum, encode it in base64 and return it as string
// The choice of sha256 for checksum is arbitrary.
func hashString(str string) string {
	hash := sha256.Sum256([]byte(str))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func unpackUser(s *schema.ResourceData) *v1.User {
	d := &ResourceData{s}
	user := new(v1.User)

	user.Name = d.getStringRef("name", false)
	user.Email = d.getStringRef("email", false)
	user.Admin = d.getBoolRef("admin", false)
	user.ProfileUpdatable = d.getBoolRef("profile_updatable", false)
	user.DisableUIAccess = d.getBoolRef("disable_ui_access", false)
	user.InternalPasswordDisabled = d.getBoolRef("internal_password_disabled", false)
	user.Groups = d.getSetRef("groups")
	user.Password = d.getStringRef("password", true)

	return user
}

func packUser(user *v1.User, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("name", user.Name))
	logErr(d.Set("email", user.Email))
	logErr(d.Set("admin", user.Admin))
	logErr(d.Set("profile_updatable", user.ProfileUpdatable))
	logErr(d.Set("disable_ui_access", user.DisableUIAccess))
	logErr(d.Set("internal_password_disabled", user.InternalPasswordDisabled))

	if user.Groups != nil {
		logErr(d.Set("groups", schema.NewSet(schema.HashString, castToInterfaceArr(*user.Groups))))
	}

	if hasErr {
		return fmt.Errorf("failed to pack user")
	}

	return nil
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	user := unpackUser(d)

	if user.Name == nil {
		return fmt.Errorf("user name cannot be nil")
	}

	if user.Password == nil {
		user.Password = artifactory.String(generatePassword())
	}

	_, err := c.V1.Security.CreateOrReplaceUser(context.Background(), *user.Name, user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		c := m.(*artifactory.Artifactory)
		_, resp, err := c.V1.Security.GetUser(context.Background(), d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error describing user: %s", err))
		}

		if resp.StatusCode == http.StatusNotFound {
			return resource.RetryableError(fmt.Errorf("expected user to be created, but currently not found"))
		}

		return resource.NonRetryableError(resourceUserRead(d, m))
	})
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	user, resp, err := c.V1.Security.GetUser(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packUser(user, d)
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	user := unpackUser(d)
	if user.Password != nil && len(*user.Password) == 0 {
		user.Password = nil
	}

	_, err := c.V1.Security.UpdateUser(context.Background(), d.Id(), user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resourceUserRead(d, m)
}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)
	user := unpackUser(d)
	_, resp, err := c.V1.Security.DeleteUser(context.Background(), *user.Name)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

// generatePassword used as default func to generate user passwords
func generatePassword() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, randomPasswordLength)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}
