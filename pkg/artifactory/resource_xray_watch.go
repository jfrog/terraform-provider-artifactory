package artifactory

import (
	"errors"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceXrayWatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceXrayWatchCreate,
		Read:   resourceXrayWatchRead,
		Delete: resourceXrayWatchDelete,
		DeprecationMessage: "Xray resources will be removed from this provider on or after March 31, 2022. " +
			"Please use the separate Terraform Provider Xray: https://github.com/jfrog/terraform-provider-xray. " +
			"Terraform Provider Registry link: https://registry.terraform.io/providers/jfrog/xray",
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bin_mgr_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"repo_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local",
								"remote",
							}, false),
						},
						"filters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									// TODO this can be either a string or possibly a json blob
									// eg "value":{"ExcludePatterns":[],"IncludePatterns":["*"]}
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"assigned_policies": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"watch_recipients": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceXrayWatchCreate(d *schema.ResourceData, m interface{}) error {
	return errors.New("Use Xray provider (https://github.com/jfrog/terraform-provider-xray) resource instead. Also see https://www.terraform.io/plugin/sdkv2/best-practices/deprecations#provider-data-source-or-resource-removal for resource removal process.")
}

func resourceXrayWatchRead(d *schema.ResourceData, m interface{}) error {
	return errors.New("Use Xray provider (https://github.com/jfrog/terraform-provider-xray) resource instead. Also see https://www.terraform.io/plugin/sdkv2/best-practices/deprecations#provider-data-source-or-resource-removal for resource removal process.")
}

func resourceXrayWatchDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete("xray/api/v2/watches/" + d.Id())
	return err
}
