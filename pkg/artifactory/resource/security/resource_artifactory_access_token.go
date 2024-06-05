package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceArtifactoryAccessToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessTokenCreate,
		ReadContext:   resourceAccessTokenRead,
		DeleteContext: resourceAccessTokenDelete,

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
				AtLeastOneOf: []string{"end_date", "end_date_relative"},
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
				AtLeastOneOf: []string{"end_date", "end_date_relative"},
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

		DeprecationMessage: "This resource is being deprecated and replaced by artifactory_scoped_token",
	}
}

func resourceAccessTokenCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_access_token deprecated. Use artifactory_scoped_token instead")
}

func resourceAccessTokenRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_access_token deprecated. Use artifactory_scoped_token instead")
}

func resourceAccessTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_access_token deprecated. Use artifactory_scoped_token instead")
}
