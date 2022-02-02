package xray

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceXrayLicensePolicyV2() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: resourceXrayPolicyCreate,
		ReadContext:   resourceXrayPolicyRead,
		UpdateContext: resourceXrayPolicyUpdate,
		DeleteContext: resourceXrayPolicyDelete,
		Description: "Creates an xray policy using V2 of the underlying APIs. Please note: " +
			"It's only compatible with Bearer token auth method (Identity and Access => Access Tokens",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the policy (must be unique)",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "More verbose description of the policy",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Type of the policy",
				ValidateDiagFunc: inList("Security", "License"),
			},
			"author": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User, who created the policy",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"modified": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Modification timestamp",
			},
			"rule": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Nested block describing a policy rule. Described below",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Name of the rule",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
						},
						"priority": {
							Type:             schema.TypeInt,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
							Description:      "Integer describing the rule priority. Must be at least 1",
						},
						"criteria": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Nested block describing the criteria for the policy. Described below",
							MinItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"banned_licenses": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "A list of OSS license names that may not be attached to a component.",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: licenseTypeValidator,
										},
									},
									"allowed_licenses": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "A list of OSS license names that may be attached to a component.",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: licenseTypeValidator,
										},
									},
									"allow_unknown": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "A violation will be generated for artifacts with unknown licenses (`true` or `false`).",
									},
									"multi_license_permissive": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Do not generate a violation if at least one license is valid in cases whereby multiple licenses were detected on the component",
									},
								},
							},
						},
						"actions": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Nested block describing the actions to be applied by the policy. Described below.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"webhooks": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "A list of Xray-configured webhook URLs to be invoked if a violation is triggered.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"mails": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "A list of email addressed that will get emailed when a violation is triggered.",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateIsEmail,
										},
									},
									"block_download": {
										Type:        schema.TypeSet,
										Required:    true,
										Description: "Nested block describing artifacts that should be blocked for download if a violation is triggered. Described below.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unscanned": {
													Type:        schema.TypeBool,
													Required:    true,
													Description: " Whether or not to block download of artifacts that meet the artifact `filters` for the associated `xray_watch` resource but have not been scanned yet.",
												},
												"active": {
													Type:        schema.TypeBool,
													Required:    true,
													Description: "Whether or not to block download of artifacts that meet the artifact and severity `filters` for the associated `xray_watch` resource.",
												},
											},
										},
									},
									"block_release_bundle_distribution": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Blocks Release Bundle distribution to Edge nodes if a violation is found.",
									},
									"fail_build": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether or not the related CI build should be marked as failed if a violation is triggered. This option is only available when the policy is applied to an `xray_watch` resource with a `type` of `builds`.",
									},
									"notify_deployer": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Sends an email message to component deployer with details about the generated Violations.",
									},
									"notify_watch_recipients": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Sends an email message to all configured recipients inside a specific watch with details about the generated Violations.",
									},
									"create_ticket_enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Create Jira Ticket for this Policy Violation. Requires configured Jira integration.",
									},
									"custom_severity": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          "High",
										Description:      "The severity of violation to be triggered if the `criteria` are met.",
										ValidateDiagFunc: inList("Critical", "High", "Medium", "Low"),
									},
									"build_failure_grace_period_in_days": {
										Type:             schema.TypeInt,
										Optional:         true,
										Description:      "Allow grace period for certain number of days. All violations will be ignored during this time. To be used only if `fail_build` is enabled.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
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
