package artifactory

import (
	"errors"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceXrayPolicy() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Create:        resourceXrayPolicyCreate,
		Read:          resourceXrayPolicyRead,
		Delete:        resourceXrayPolicyDelete,
		DeprecationMessage: "Xray resources will be removed from this provider on or after March 31, 2022. " +
			"Please use the separate Terraform Provider Xray: https://github.com/jfrog/terraform-provider-xray. " +
			"Terraform Provider Registry link: https://registry.terraform.io/providers/jfrog/xray",
		Description: "Creates an xray policy using V1 of the underlying APIs. Please note: " +
			"It's only compatible with Bearer token auth method (Identity and Access => Access Tokens",

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"criteria": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Security criteria
									"min_severity": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"cvss_range": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from": {
													Type:     schema.TypeInt, // Yes, the xray web ui allows floats. The go library says ints. :(
													Required: true,
												},
												"to": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									// License Criteria
									"allow_unknown": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"banned_licenses": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"allowed_licenses": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"actions": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mails": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"fail_build": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"block_download": {
										Type:     schema.TypeList,
										Required: true,
										// TODO: In an ideal world, this would be optional (see note in expandActions)
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unscanned": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"active": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"webhooks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"custom_severity": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceXrayPolicyCreate(d *schema.ResourceData, m interface{}) error {
	return errors.New("Use Xray provider resource instead")
}

func resourceXrayPolicyRead(d *schema.ResourceData, m interface{}) error {
	return errors.New("Use Xray provider resource instead")
}

func resourceXrayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete("xray/api/v1/policies/" + d.Id())
	return err
}
