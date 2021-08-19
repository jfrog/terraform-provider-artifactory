package artifactory

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var warning = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

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
	c := m.(*ArtClient).ArtOld

	user := unpackUser(d)

	if user.Name == nil {
		return fmt.Errorf("user name cannot be nil")
	}

	if user.Password == nil {
		warning.Println("No password supplied. One will be generated and this can fail as your RT password policy can't be known here")
		user.Password = artifactory.String(generatePassword(16))
	}

	_, err := c.V1.Security.CreateOrReplaceUser(context.Background(), *user.Name, user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		c := m.(*ArtClient).ArtOld
		_, resp, err := c.V1.Security.GetUser(context.Background(), d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error describing user: %s", err))
		}

		if resp.StatusCode == http.StatusNotFound {
			return resource.RetryableError(fmt.Errorf("expected user to be created, but currently not found"))
		}

		return nil
	})
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	user, resp, err := c.V1.Security.GetUser(context.Background(), d.Id())

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	return packUser(user, d)
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

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
	c := m.(*ArtClient).ArtOld
	user := unpackUser(d)
	_, resp, err := c.V1.Security.DeleteUser(context.Background(), *user.Name)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("user %s not deleted. Status code: %d", *user.Name, resp.StatusCode)
}

// generatePassword used as default func to generate user passwords. It's possible for this to be incompatible with what
// rt will allow, but there is no way to know what rules are in place
func generatePassword(length int) string {
	randSelect := func(str string, count int) string {
		strLen := len(str)
		result := make([]byte, count)
		for i := range result {
			result[i] = str[rand.Intn(strLen)]
		}
		return string(result)
	}
	up := func(count int) string {
		return randSelect("ABCDEFGHIJKLMNOPQRSTUVWXYZ", count)
	}
	low := func(count int) string {
		return randSelect("abcdefghijklmnopqrstuvwxyz", count)
	}
	dig := func(count int) string {
		return randSelect("0123456789", count)
	}
	spec := func(count int) string {
		return randSelect("!@#$%^&*()-_+=[]{}|<>?/~'\"", count)
	}
	lowLen := length / 2
	runes := []rune(low(lowLen-1) + up(length-lowLen-1) + spec(1) + dig(1))

	rand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})
	return string(runes)
}
