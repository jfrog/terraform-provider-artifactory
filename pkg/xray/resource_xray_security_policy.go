package xray

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceXraySecurityPolicyV2() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Create:        resourceXrayPolicyCreate,
		Read:          resourceXrayPolicyRead,
		Update:        resourceXrayPolicyUpdate,
		Delete:        resourceXrayPolicyDelete,
		Description: "Creates an xray policy using V2 of the underlying APIs. Please note: " +
			"It's only compatible with Bearer token auth method (Identity and Access => Access Tokens",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			// not in create policy body, but it is in the get call response.
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
			//
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
													Type:     schema.TypeFloat,
													Required: true,
												},
												"to": {
													Type:     schema.TypeFloat,
													Required: true,
												},
											},
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
									"webhooks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"mails": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"block_download": {
										Type:     schema.TypeList,
										Required: true,
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
									"block_release_bundle_distribution": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"fail_build": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"notify_deployer": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"notify_watch_recipients": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"create_ticket_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"build_failure_grace_period_in_days": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  3,
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
