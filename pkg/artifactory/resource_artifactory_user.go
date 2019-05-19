package artifactory

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/schema"
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
			},
			"profile_updatable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"disable_ui_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"internal_password_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
		},
	}
}

func unpackUser(s *schema.ResourceData) *v1.User {
	d := &ResourceData{s}
	user := new(v1.User)

	user.Name = d.getStringRef("name")
	user.Email = d.getStringRef("email")
	user.Admin = d.getBoolRef("admin")
	user.ProfileUpdatable = d.getBoolRef("profile_updatable")
	user.DisableUIAccess = d.getBoolRef("disable_ui_access")
	user.InternalPasswordDisabled = d.getBoolRef("internal_password_disabled")
	user.Groups = d.getSetRef("groups")

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

	if pass, ok := os.LookupEnv(fmt.Sprintf("TF_USER_%s_PASSWORD", *user.Name)); ok {
		user.Password = artifactory.String(pass)
	} else {
		nameSum := md5.Sum([]byte(*user.Name))
		if encPass, ok := os.LookupEnv(fmt.Sprintf("TF_USER_%x_PASSWORD_ENC", nameSum)); ok {
			pass, err := base64.StdEncoding.DecodeString(encPass)

			if err != nil {
				return fmt.Errorf("base64 username exists but password not encoded correctly: %s", err)
			}
			user.Password = artifactory.String(string(pass))
		} else {
			user.Password = artifactory.String(generatePassword())
		}
	}

	_, err := c.V1.Security.CreateOrReplaceUser(context.Background(), *user.Name, user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resourceUserRead(d, m)
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
