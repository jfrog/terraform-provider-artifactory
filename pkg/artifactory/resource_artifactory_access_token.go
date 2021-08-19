package artifactory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	artifactoryold "github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceArtifactoryAccessToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessTokenCreate,
		Read:   resourceAccessTokenRead,
		Delete: resourceAccessTokenDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"end_date_relative": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"end_date"},
				ValidateFunc: func(i interface{}, k string) ([]string, []error) {
					v, ok := i.(string)
					if !ok {
						return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
					}

					if strings.TrimSpace(v) == "" {
						return nil, []error{fmt.Errorf("%q must not be empty", k)}
					}

					return nil, nil
				},
			},
			"end_date": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"end_date_relative"},
				ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
					v, ok := i.(string)
					if !ok {
						errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
						return warnings, errors
					}

					if _, err := time.Parse(time.RFC3339, v); err != nil {
						errors = append(errors, fmt.Errorf("expected %q to be a valid RFC3339 date, got %q: %+v", k, i, err))
					}

					return warnings, errors
				},
			},
			"refreshable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"admin_token": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"groups"},
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"access_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"refresh_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAccessTokenCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).ArtOld
	grantType := "client_credentials" // client_credentials is the only supported type

	tokenOptions := v1.AccessTokenOptions{}
	resourceData := &ResourceData{d}

	date, expiresIn, err := getDate(d)
	if err != nil {
		return err
	}

	tokenOptions.ExpiresIn = expiresIn
	err = d.Set("end_date", date.Format(time.RFC3339))
	if err != nil {
		return err
	}

	refreshable := resourceData.Get("refreshable").(bool)
	audience := d.Get("audience").(string)

	tokenOptions.Audience = artifactory.String(audience)
	tokenOptions.GrantType = &grantType
	tokenOptions.Refreshable = artifactory.String(strconv.FormatBool(refreshable))
	tokenOptions.Username = resourceData.getStringRef("username", false)

	username := resourceData.Get("username").(string)
	userExists, _ := checkUserExists(client, username)

	if !userExists && len(resourceData.Get("groups").([]interface{})) == 0 {
		return fmt.Errorf("you must specify at least 1 group when creating a token for a non-existant user - %s, or correct the username", username)
	}

	err = unpackGroups(d, client, &tokenOptions)
	if err != nil {
		return err
	}

	err = unpackAdminToken(d, &tokenOptions)
	if err != nil {
		return err
	}

	accessToken, _, err := client.V1.Security.CreateToken(context.Background(), &tokenOptions)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(hashcode.String(*accessToken.AccessToken)))

	err = d.Set("access_token", *accessToken.AccessToken)
	if err != nil {
		return err
	}

	refreshToken := ""
	if refreshable {
		refreshToken = *accessToken.RefreshToken
	}

	err = d.Set("refresh_token", refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func resourceAccessTokenRead(_ *schema.ResourceData, _ interface{}) error {
	// Terraform requires that the read function is always implemented.
	// However, Artifactory does not have an API to read a token.
	return nil
}

func resourceAccessTokenDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).ArtOld

	// Artifactory only allows you to revoke a token if the there is no expiry.
	// Otherwise, Artifactory will ensure the token is revoked at the expiry time.
	// https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-ViewingandRevokingTokens
	// https://www.jfrog.com/jira/browse/RTFACT-15293

	// If relative end date is empty, then a fixed end date was set
	// Therefore, Artifactory will expire the token automatically
	endDateRelative := d.Get("end_date_relative").(string)
	if endDateRelative == "" {
		log.Printf("[DEBUG] Token is not revoked. It will expire at " + d.Get("end_date").(string))
		return nil
	}

	// Convert end date relative to duration in seconds
	duration, err := time.ParseDuration(endDateRelative)
	if err != nil {
		return fmt.Errorf("unable to parse `end_date_relative` (%s) as a duration", endDateRelative)
	}

	// If the token has no duration, it does not expire.
	// Therefore revoke the token.
	if duration.Seconds() == 0 {
		log.Printf("[DEBUG] Revoking token")
		revokeOptions := v1.AccessTokenRevokeOptions{}
		revokeOptions.Token = d.Get("access_token").(string)

		_, resp, err := client.V1.Security.RevokeToken(context.Background(), revokeOptions)

		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] Token Revoked")
			return nil
		}
		return err
	}

	// If the duration is set, Artifactory will automatically revoke the token.
	log.Printf("[DEBUG] Token is not revoked. It will expire at " + d.Get("end_date").(string))

	return nil
}

func unpackGroups(d *schema.ResourceData, client *artifactoryold.Artifactory, tokenOptions *v1.AccessTokenOptions) error {
	if srcGroups, ok := d.GetOk("groups"); ok {
		groups := make([]string, len(srcGroups.([]interface{})))
		for i, group := range srcGroups.([]interface{}) {
			groups[i] = group.(string)

			if exist, err := checkGroupExists(client, groups[i]); !exist {
				return err
			}
		}

		scopedGroupString := strings.Join(groups[:], ",")
		scope := "member-of-groups:\"" + scopedGroupString + "\""
		tokenOptions.Scope = &scope
	}

	return nil
}

func unpackAdminToken(d *schema.ResourceData, tokenOptions *v1.AccessTokenOptions) error {
	if adminToken, ok := d.GetOk("admin_token"); ok {
		set := adminToken.(*schema.Set)
		val := set.List()[0].(map[string]interface{})

		instanceID := val["instance_id"].(string)

		scope := "jfrt@" + instanceID + ":admin"
		tokenOptions.Scope = &scope
	}

	return nil
}

func checkUserExists(client *artifactoryold.Artifactory, name string) (bool, error) {
	_, resp, err := client.V1.Security.GetUser(context.Background(), name)
	if err != nil {
		return false, err
	}
	// If there is an error, it possible the user does not exist.
	if resp != nil {
		// Check if the user does not exist in artifactory
		if resp.StatusCode == http.StatusNotFound {
			return false, errors.New("user must exist in artifactory")
		}

		// If we cannot search for Users, the current user is not an admin
		// So, we'll let this through and let the CreateToken function error if there is a misconfiguration.
		if resp.StatusCode == http.StatusForbidden {
			return true, nil
		}
	}
	return false, err
}

func checkGroupExists(client *artifactoryold.Artifactory, name string) (bool, error) {
	_, resp, err := client.V1.Security.GetGroup(context.Background(), name)

	// If there is an error, it possible the group does not exist.
	if err != nil {
		if resp != nil {
			// Check if the group does not exist in artifactory
			if resp.StatusCode == http.StatusNotFound {
				return false, errors.New("group must exist in artifactory")
			}

			// If we cannot search for groups, the current user is not an admin and they can only specify groups they belong to.
			// Therefore, we return true and rely on Artifactory to error if the user has specified a wrong group.
			if resp.StatusCode == http.StatusForbidden {
				return true, nil
			}
		}

		return false, err
	}

	return true, nil
}

// Inspired by azure ad implementation
func getDate(d *schema.ResourceData) (*time.Time, *int, error) {
	var endDate time.Time
	now := time.Now()

	if v := d.Get("end_date").(string); v != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse the provided end date %q: %+v", v, err)
		}
	} else if v := d.Get("end_date_relative").(string); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse `end_date_relative` %s as a duration", v)
		}
		// Artifactory's minimum duration is in seconds.
		// The consumer should either specify 0 for a non-expiring token, or >= 1 seconds
		// If the consumer passes in configuration that is between 0 and 1 seconds, they have used a smaller time unit that seconds.
		if d.Nanoseconds() > 0 && d.Seconds() < 1 {
			return nil, nil, fmt.Errorf("minimum duration is 1 second, but `end_date_relative` is %s", v)
		}
		endDate = time.Now().Add(d)
	} else {
		return nil, nil, fmt.Errorf("one of `end_date` or `end_date_relative` must be specified")
	}

	differenceInSeconds := int(endDate.Sub(now).Seconds())

	if differenceInSeconds < 0 {
		return nil, nil, fmt.Errorf("end date must be in the future, but is %s", endDate.String())
	}
	return &endDate, &differenceInSeconds, nil
}
