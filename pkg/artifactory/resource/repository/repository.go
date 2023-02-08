package repository

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/exp/slices"
)

var CompressionFormats = map[string]*schema.Schema{
	"index_compression_formats": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Set:      schema.HashString,
		Optional: true,
	},
}

type ContentSynchronisation struct {
	Enabled    bool                             `json:"enabled"`
	Statistics ContentSynchronisationStatistics `json:"statistics"`
	Properties ContentSynchronisationProperties `json:"properties"`
	Source     ContentSynchronisationSource     `json:"source"`
}

type ContentSynchronisationStatistics struct {
	Enabled bool `hcl:"statistics_enabled" json:"enabled"`
}

type ContentSynchronisationProperties struct {
	Enabled bool `hcl:"properties_enabled" json:"enabled"`
}

type ContentSynchronisationSource struct {
	OriginAbsenceDetection bool `hcl:"source_origin_absence_detection" json:"originAbsenceDetection"`
}

type ReadFunc func(d *schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, resty gets confused and thinks it's a map
type Constructor func() (interface{}, error)

func mkRepoCreate(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().
			AddRetryCondition(client.RetryOnMergeError).
			SetBody(repo).
			SetPathParam("key", key).
			Put(RepositoriesEndpoint)

		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(key)
		return read(ctx, d, m)
	}
}

func MkRepoRead(pack packer.PackFunc, construct Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, err := construct()
		if err != nil {
			return diag.FromErr(err)
		}

		// repo must be a pointer
		resp, err := m.(*resty.Client).R().
			SetResult(repo).
			SetPathParam("key", d.Id()).
			Get(RepositoriesEndpoint)

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.FromErr(pack(repo, d))
	}
}

func mkRepoUpdate(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			AddRetryCondition(client.RetryOnMergeError).
			SetBody(repo).
			SetPathParam("key", d.Id()).
			Post(RepositoriesEndpoint)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(key)

		projectKeyChanged := d.HasChange("project_key")
		tflog.Debug(ctx, fmt.Sprintf("projectKeyChanged: %v", projectKeyChanged))
		if projectKeyChanged {
			old, new := d.GetChange("project_key")
			oldProjectKey := old.(string)
			newProjectKey := new.(string)
			tflog.Debug(ctx, fmt.Sprintf("oldProjectKey: %v, newProjectKey: %v", oldProjectKey, newProjectKey))

			assignToProject := len(oldProjectKey) == 0 && len(newProjectKey) > 0
			unassignFromProject := len(oldProjectKey) > 0 && len(newProjectKey) == 0
			tflog.Debug(ctx, fmt.Sprintf("assignToProject: %v, unassignFromProject: %v", assignToProject, unassignFromProject))

			var err error
			if assignToProject {
				err = assignRepoToProject(key, newProjectKey, m.(*resty.Client))
			} else if unassignFromProject {
				err = unassignRepoFromProject(key, m.(*resty.Client))
			}

			if err != nil {
				return diag.FromErr(err)
			}
		}

		return read(ctx, d, m)
	}
}

func assignRepoToProject(repoKey string, projectKey string, client *resty.Client) error {
	_, err := client.R().
		SetPathParams(map[string]string{
			"repoKey":    repoKey,
			"projectKey": projectKey,
		}).
		Put("access/api/v1/projects/_/attach/repositories/{repoKey}/{projectKey}")
	return err
}

func unassignRepoFromProject(repoKey string, client *resty.Client) error {
	_, err := client.R().
		SetPathParam("repoKey", repoKey).
		Delete("access/api/v1/projects/_/attach/repositories/{repoKey}")
	return err
}

func deleteRepo(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*resty.Client).R().
		AddRetryCondition(client.RetryOnMergeError).
		SetPathParam("key", d.Id()).
		Delete(RepositoriesEndpoint)

	if err != nil && (resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound)) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

func Retry400(response *resty.Response, _ error) bool {
	return response.StatusCode() == http.StatusBadRequest
}

func repoExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := CheckRepo(d.Id(), m.(*resty.Client).R().AddRetryCondition(Retry400))
	return err == nil, err
}

var repoTypeValidator = validation.StringInSlice(RepoTypesSupported, false)

var RepoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()+={}[]:;<>,/?~`|\\"),
)

var RepoTypesSupported = []string{
	"alpine",
	"bower",
	"cargo",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"p2",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
	"vcs",
}

var GradleLikeRepoTypes = []string{
	"gradle",
	"sbt",
	"ivy",
}

var ProjectEnvironmentsSupported = []string{"DEV", "PROD"}

func RepoLayoutRefSchema(repositoryType string, packageType string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo_layout_ref": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: GetDefaultRepoLayoutRef(repositoryType, packageType),
			Description: "Repository layout key for the local repository",
		},
	}
}

// HandleResetWithNonExistentValue Special handling for field that requires non-existant value for RT
//
// Artifactory REST API will not accept empty string or null to reset value to not set
// Instead, using a non-existant value works as a workaround
// To ensure we don't accidentally set the value to a valid value, we use a UUID v4 string
func HandleResetWithNonExistentValue(d *util.ResourceData, key string) string {
	value := d.GetString(key, false)

	// When value has changed and is empty string, then it has been removed from
	// the Terraform configuration.
	if value == "" && d.HasChange(key) {
		return fmt.Sprintf("non-existant-value-%d", test.RandomInt())
	}

	return value
}

func projectEnvironmentsDiff(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if data, ok := diff.GetOk("project_environments"); ok {
		projectEnvironments := data.(*schema.Set).List()

		for _, projectEnvironment := range projectEnvironments {
			if !slices.Contains(ProjectEnvironmentsSupported, projectEnvironment.(string)) {
				return fmt.Errorf("project_environment %s not allowed", projectEnvironment)
			}
		}
	}

	return nil
}

func MkResourceSchema(skeema map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor Constructor) *schema.Resource {
	var reader = MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: mkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: mkRepoUpdate(unpack, reader),
		DeleteContext: deleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        skeema,
		CustomizeDiff: projectEnvironmentsDiff,
	}
}

const RepositoriesEndpoint = "artifactory/api/repositories/{key}"

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.SetPathParam("key", id).Head(RepositoriesEndpoint)
}

func ValidateRepoLayoutRefSchemaOverride(_ interface{}, _ cty.Path) diag.Diagnostics {
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Always override repo_layout_ref attribute in the schema",
			Detail:   "Always override repo_layout_ref attribute in the schema on top of base schema",
		},
	}
}

type SupportedRepoClasses struct {
	RepoLayoutRef      string
	SupportedRepoTypes map[string]bool
}

// GetDefaultRepoLayoutRef return the default repo layout by Repository Type & Package Type
func GetDefaultRepoLayoutRef(repositoryType string, packageType string) func() (interface{}, error) {
	return func() (interface{}, error) {
		if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
			return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
		}
		return "", fmt.Errorf("default repo layout not found for repository type %v & package type %v", repositoryType, packageType)
	}
}
