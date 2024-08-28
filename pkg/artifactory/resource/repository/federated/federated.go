package federated

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

const rclass = "federated"
const RepositoriesEndpoint = "artifactory/api/repositories/{key}"

var PackageTypesLikeGeneric = []string{
	"bower",
	"chef",
	"cocoapods",
	"composer",
	"conda",
	"cran",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"helm",
	"npm",
	"opkg",
	"puppet",
	"pypi",
	"swift",
	"vagrant",
}

type RepoParams struct {
	Proxy        string `json:"proxy"`
	DisableProxy bool   `json:"disableProxy"`
}

type Member struct {
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

var SchemaGeneratorV3 = func(isRequired bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		repository.ProxySchema,
		map[string]*schema.Schema{
			"cleanup_on_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete all federated members on `terraform destroy` if set to `true`. Caution: it will delete all the repositories in the federation on other Artifactory instances.",
			},
			"member": {
				Type:     schema.TypeSet,
				Required: isRequired,
				Optional: !isRequired,
				Description: "The list of Federated members. If a Federated member receives a request that does not include the repository URL, it will " +
					"automatically be added with the combination of the configured base URL and `key` field value. " +
					"Note that each of the federated members will need to have a base URL set. Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)" +
					" to set up Federated repositories correctly.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Full URL to ending with the repositoryName",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "Represents the active state of the federated member. It is supported to " +
								"change the enabled status of my own member. The config will be updated on the other " +
								"federated members automatically.",
						},
					},
				},
			},
		},
	)
}

var federatedSchemaV3 = SchemaGeneratorV3(true)

var SchemaGeneratorV4 = func(isRequired bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		federatedSchemaV3,
		map[string]*schema.Schema{
			"cleanup_on_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete all federated members on `terraform destroy` if set to `true`. Caution: it will delete all the repositories in the federation on other Artifactory instances. Set `access_token` attribute if Access Federation for access tokens is not enabled.",
			},
			"member": {
				Type:     schema.TypeSet,
				Required: isRequired,
				Optional: !isRequired,
				Description: "The list of Federated members. If a Federated member receives a request that does not include the repository URL, it will " +
					"automatically be added with the combination of the configured base URL and `key` field value. " +
					"Note that each of the federated members will need to have a base URL set. Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)" +
					" to set up Federated repositories correctly.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Full URL to ending with the repositoryName",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "Represents the active state of the federated member. It is supported to " +
								"change the enabled status of my own member. The config will be updated on the other " +
								"federated members automatically.",
						},
						"access_token": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
							Description:      "Admin access token for this member Artifactory instance. Used in conjunction with `cleanup_on_delete` attribute when Access Federation for access tokens is not enabled.",
						},
					},
				},
			},
		},
	)
}

var federatedSchemaV4 = SchemaGeneratorV4(true)

func unpackMembers(data *schema.ResourceData) []Member {
	d := &utilsdk.ResourceData{ResourceData: data}
	var members []Member

	if v, ok := d.GetOk("member"); ok {
		federatedMembers := v.(*schema.Set).List()
		if len(federatedMembers) == 0 {
			return members
		}

		for _, federatedMember := range federatedMembers {
			id := federatedMember.(map[string]interface{})

			member := Member{
				Url:     id["url"].(string),
				Enabled: id["enabled"].(bool),
			}
			members = append(members, member)
		}
	}
	return members
}

func unpackRepoParams(data *schema.ResourceData) RepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}

	return RepoParams{
		Proxy:        d.GetString("proxy", false),
		DisableProxy: d.GetBool("disable_proxy", false),
	}
}

func PackMembers(members []Member, d *schema.ResourceData) error {
	setValue := utilsdk.MkLens(d)

	var federatedMembers []interface{}

	for _, member := range members {
		federatedMember := map[string]interface{}{
			"url":          member.Url,
			"enabled":      member.Enabled,
			"access_token": nil,
		}

		// find matching member to restore the 'access_token' value
		if v, ok := d.GetOk("member"); ok {
			matchedMember, found := lo.Find(
				v.(*schema.Set).List(),
				func(m interface{}) bool {
					id := m.(map[string]interface{})
					return id["url"] == member.Url
				},
			)

			if found {
				id := matchedMember.(map[string]interface{})
				if v, ok := id["access_token"]; ok && v != "" {
					federatedMember["access_token"] = v.(string)
				}
			}
		}

		federatedMembers = append(federatedMembers, federatedMember)
	}

	errors := setValue("member", federatedMembers)
	if len(errors) > 0 {
		return fmt.Errorf("failed saving members to state %q", errors)
	}

	return nil
}

func configSync(ctx context.Context, repoKey string, m interface{}) diag.Diagnostics {
	var ds diag.Diagnostics

	tflog.Info(ctx,
		"triggering synchronization of the federated member configuration",
		map[string]interface{}{
			"repoKey": repoKey,
		},
	)
	resp, restErr := m.(util.ProviderMetadata).Client.R().
		SetPathParam("repositoryKey", repoKey).
		Post("artifactory/api/federation/configSync/{repositoryKey}")
	if restErr != nil {
		ds = append(ds, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "failed to trigger synchronization of the federated member configuration",
			Detail:   restErr.Error(),
		})
	}
	if resp.IsError() {
		ds = append(ds, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "failed to trigger synchronization of the federated member configuration",
			Detail:   resp.String(),
		})
	}

	return ds
}

func createRepo(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		var ds diag.Diagnostics
		ds = append(ds, repository.Create(ctx, d, m, unpack)...)
		if ds.HasError() {
			return ds
		}

		ds = append(ds, configSync(ctx, d.Id(), m)...)
		if ds.HasError() {
			return ds
		}

		return append(ds, read(ctx, d, m)...)
	}
}

func updateRepo(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		var ds diag.Diagnostics
		ds = append(ds, repository.Update(ctx, d, m, unpack)...)
		if ds.HasError() {
			return ds
		}

		ds = append(ds, configSync(ctx, d.Id(), m)...)
		if ds.HasError() {
			return ds
		}

		return append(ds, read(ctx, d, m)...)
	}
}

func deleteRepo(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ds := diag.Diagnostics{}

	restyClient := m.(util.ProviderMetadata).Client

	// For federated repositories we delete all the federated members (except the initial repo member), if the flag `cleanup_on_delete` is set to `true`
	s := &utilsdk.ResourceData{ResourceData: d}
	initialRepoName := s.GetString("key", false)

	if v, ok := d.GetOk("member"); ok && s.GetBool("cleanup_on_delete", false) {
		federatedMembers := v.(*schema.Set).List()

		for _, federatedMember := range federatedMembers {
			id := federatedMember.(map[string]interface{})

			memberUrl := id["url"].(string) // example "https://artifactory-instance.com/artifactory/federated-generic-repository-example"
			parsedMemberUrl, err := url.Parse(memberUrl)
			if err != nil {
				return diag.FromErr(err)
			}

			memberHost := memberUrl[:strings.Index(memberUrl, parsedMemberUrl.Path)]
			memberRepoName := strings.ReplaceAll(memberUrl, memberUrl[:strings.LastIndex(memberUrl, "/")+1], "")

			if initialRepoName != memberRepoName || !strings.HasPrefix(memberUrl, restyClient.BaseURL) {
				request := restyClient.R().
					AddRetryCondition(client.RetryOnMergeError).
					SetPathParam("key", memberRepoName)

				accessToken := ""
				if v, ok := id["access_token"]; ok {
					accessToken = v.(string)
				}

				if accessToken != "" {
					request.SetAuthToken(accessToken)
				}

				memberAPIURL := fmt.Sprintf("%s/%s", memberHost, RepositoriesEndpoint)
				resp, err := request.Delete(memberAPIURL)

				if err != nil {
					ds = append(
						ds,
						diag.Diagnostic{
							Severity: diag.Warning,
							Summary:  "Failed to delete federated repository member",
							Detail:   fmt.Sprintf("Error deleting member repository %s: %s", memberRepoName, err.Error()),
						},
					)
				}

				if resp.IsError() {
					ds = append(
						ds,
						diag.Diagnostic{
							Severity: diag.Warning,
							Summary:  "Failed to delete federated repository member",
							Detail:   fmt.Sprintf("Error deleting member repository %s: %s", memberRepoName, resp.String()),
						},
					)
				}
			}
		}
	}

	resp, err := restyClient.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetPathParam("key", d.Id()).
		Delete(RepositoriesEndpoint)

	if err != nil {
		ds = append(
			ds,
			diag.FromErr(err)...,
		)
		return ds
	}

	if resp.IsError() {
		ds = append(
			ds,
			diag.Errorf("%s", resp.String())...,
		)
		return ds
	}

	d.SetId("")

	return ds
}

func mkResourceSchema(skeema map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor repository.Constructor) *schema.Resource {
	var reader = repository.MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: createRepo(unpack, reader),
		ReadContext:   reader,
		UpdateContext: updateRepo(unpack, reader),
		DeleteContext: deleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        skeema,
		SchemaVersion: 4,
		StateUpgraders: []schema.StateUpgrader{
			{
				// this only works because the schema hasn't changed, except the removal of default value
				// from `project_key` attribute.
				Type:    resourceV2(federatedSchemaV3).CoreConfigSchema().ImpliedType(),
				Upgrade: repository.ResourceUpgradeProjectKey,
				Version: 2,
			},
			{
				Type:    resourceV3().CoreConfigSchema().ImpliedType(),
				Upgrade: upgradeMemberAccessToken,
				Version: 3,
			},
		},
		CustomizeDiff: customdiff.All(
			repository.ProjectEnvironmentsDiff,
			repository.VerifyDisableProxy,
		),
	}
}

func resourceV2(skeema map[string]*schema.Schema) *schema.Resource {
	return &schema.Resource{
		Schema: skeema,
	}
}

func resourceV3() *schema.Resource {
	return &schema.Resource{
		Schema: federatedSchemaV3,
	}
}

func upgradeMemberAccessToken(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if v, ok := rawState["member"]; ok {
		for _, m := range v.([]interface{}) {
			id := m.(map[string]interface{})
			id["access_token"] = nil
		}
	}

	return rawState, nil
}
