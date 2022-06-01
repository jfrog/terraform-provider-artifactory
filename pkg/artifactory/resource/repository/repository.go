package repository

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/test"
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
type Constructor func() interface{}

// UnpackFunc must return a pointer to a struct and the resource id
type UnpackFunc func(s *schema.ResourceData) (interface{}, string, error)

type PackFunc func(repo interface{}, d *schema.ResourceData) error

func mkRepoCreate(unpack UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().
			AddRetryCondition(client.RetryOnMergeError).
			SetBody(repo).
			Put(RepositoriesEndpoint + key)

		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(key)
		return read(ctx, d, m)
	}
}

func mkRepoRead(pack PackFunc, construct Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo := construct()
		// repo must be a pointer
		resp, err := m.(*resty.Client).R().SetResult(repo).Get(RepositoriesEndpoint + d.Id())

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

func mkRepoUpdate(unpack UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().
			AddRetryCondition(client.RetryOnMergeError).
			SetBody(repo).
			Post(RepositoriesEndpoint + d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(key)
		return read(ctx, d, m)
	}
}

func deleteRepo(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*resty.Client).R().
		AddRetryCondition(client.RetryOnMergeError).
		Delete(RepositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound)) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

func Retry400(response *resty.Response, err error) bool {
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

// Special handling for field that requires non-existant value for RT
//
// Artifactory REST API will not accept empty string or null to reset value to not set
// Instead, using a non-existant value works as a workaround
// To ensure we don't accidentally set the value to a valid value, we use a UUID v4 string
func HandleResetWithNonExistantValue(d *util.ResourceData, key string) string {
	value := d.GetString(key, false)

	// When value has changed and is empty string, then it has been removed from
	// the Terraform configuration.
	if value == "" && d.HasChange(key) {
		return fmt.Sprintf("non-existant-value-%d", test.RandomInt())
	}

	return value
}

// TODO universalUnpack - implement me
// func universalUnpack(payload reflect.Type, s *schema.ResourceData) (interface{}, string, error) {
// 	d := &util.ResourceData{s}
// 	var t = reflect.TypeOf(payload)
// 	var v = reflect.ValueOf(payload)
// 	if t.Kind() == reflect.Ptr {
// 		t = t.Elem()
// 		v = v.Elem()
// 	}
//
// 	for i := 0; i < t.NumField(); i++ {
// 		thing := v.Field(i)
//
// 		switch thing.Kind() {
// 		case reflect.String:
// 			v.SetString(thing.String())
// 		case reflect.Int:
// 			v.SetInt(thing.Int())
// 		case reflect.Bool:
// 			v.SetBool(thing.Bool())
// 		}
// 	}
// 	result := KeyPairPayLoad{
// 		PairName:    d.GetString("pair_name", false),
// 		PairType:    d.GetString("pair_type", false),
// 		Alias:       d.GetString("alias", false),
// 		PrivateKey:  strings.ReplaceAll(d.GetString("private_key", false), "\t", ""),
// 		PublicKey:   strings.ReplaceAll(d.GetString("public_key", false), "\t", ""),
// 		Unavailable: d.GetBool("unavailable", false),
// 	}
// 	return &result, result.PairName, nil
// }

type AutoMapper func(field reflect.StructField, thing reflect.Value) map[string]interface{}

func checkForHcl(mapper AutoMapper) AutoMapper {
	return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
		if field.Tag.Get("hcl") != "" {
			return mapper(field, thing)
		}
		return map[string]interface{}{}
	}
}

func findInspector(kind reflect.Kind) AutoMapper {
	switch kind {
	case reflect.Struct:
		return func(f reflect.StructField, t reflect.Value) map[string]interface{} {
			return lookup(t.Interface(), nil)
		}
	case reflect.Ptr:
		return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
			deref := reflect.Indirect(thing)
			if deref.CanAddr() {
				result := deref.Interface()
				if deref.Kind() == reflect.Struct {
					result = []interface{}{lookup(deref.Interface(), nil)}
				}
				return map[string]interface{}{
					fieldToHcl(field): result,
				}
			}
			return map[string]interface{}{}
		}
	case reflect.Slice:
		return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
			return map[string]interface{}{
				fieldToHcl(field): util.CastToInterfaceArr(thing.Interface().([]string)),
			}
		}
	}
	return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
		return map[string]interface{}{
			fieldToHcl(field): thing.Interface(),
		}
	}
}

// fieldToHcl this function is meant to use the HCL provided in the tag, or create a snake_case from the field name
// it actually works as expected, but dynamically working with these names was catching edge cases everywhere and
// it was/is a time sink to catch.
func fieldToHcl(field reflect.StructField) string {

	if field.Tag.Get("hcl") != "" {
		return field.Tag.Get("hcl")
	}
	var lowerFields []string
	rgx := regexp.MustCompile("([A-Z][a-z]+)")
	fields := rgx.FindAllStringSubmatch(field.Name, -1)
	for _, matches := range fields {
		for _, match := range matches[1:] {
			lowerFields = append(lowerFields, strings.ToLower(match))
		}
	}
	result := strings.Join(lowerFields, "_")
	return result
}

func lookup(payload interface{}, predicate util.HclPredicate) map[string]interface{} {

	if predicate == nil {
		predicate = allowAllPredicate
	}

	values := map[string]interface{}{}
	var t = reflect.TypeOf(payload)
	var v = reflect.ValueOf(payload)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		thing := v.Field(i)

		shouldLookup := true
		if thing.Kind() != reflect.Struct {
			hcl := fieldToHcl(field)
			shouldLookup = predicate(hcl)
		}

		if shouldLookup {
			typeInspector := findInspector(thing.Kind())
			for key, value := range typeInspector(field, thing) {
				if _, ok := values[key]; !ok {
					values[key] = value
				}
			}
		}
	}
	return values
}

func anyuHclPredicate(predicates ...util.HclPredicate) util.HclPredicate {
	return func(hcl string) bool {
		for _, predicate := range predicates {
			if predicate(hcl) {
				return true
			}
		}
		return false
	}
}

func AllHclPredicate(predicates ...util.HclPredicate) util.HclPredicate {
	return func(hcl string) bool {
		for _, predicate := range predicates {
			if !predicate(hcl) {
				return false
			}
		}
		return true
	}
}

var noClass = IgnoreHclPredicate("class", "rclass")
var NoPassword = IgnoreHclPredicate("class", "rclass", "password")

var allowAllPredicate = func(hcl string) bool {
	return true
}

func IgnoreHclPredicate(names ...string) util.HclPredicate {
	set := map[string]interface{}{}
	for _, name := range names {
		set[name] = nil
	}
	return func(hcl string) bool {
		_, found := set[hcl]
		return !found
	}
}

func ComposePacker(packers ...PackFunc) PackFunc {
	return func(repo interface{}, d *schema.ResourceData) error {
		var errors []error

		for _, packer := range packers {
			err := packer(repo, d)
			if err != nil {
				errors = append(errors, err)
			}
		}
		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving state %q", errors)
		}
		return nil
	}
}

func DefaultPacker(skeema map[string]*schema.Schema) PackFunc {
	return UniversalPack(AllHclPredicate(util.SchemaHasKey(skeema), NoPassword))
}

// UniversalPack consider making this a function that takes a predicate of what to include and returns
// a function that does the job. This would allow for the legacy code to specify which keys to keep and not
func UniversalPack(predicate util.HclPredicate) PackFunc {

	return func(payload interface{}, d *schema.ResourceData) error {
		setValue := util.MkLens(d)

		var errors []error

		values := lookup(payload, predicate)

		for hcl, value := range values {
			if predicate != nil && predicate(hcl) {
				errors = setValue(hcl, value)
			}
		}

		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving state %q", errors)
		}
		return nil
	}
}

func projectEnvironmentsDiff(_ context.Context, diff *schema.ResourceDiff, i interface{}) error {
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

func MkResourceSchema(skeema map[string]*schema.Schema, packer PackFunc, unpack UnpackFunc, constructor Constructor) *schema.Resource {
	var reader = mkRepoRead(packer, constructor)
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

// selectRandomFromMapOfStrings returns random string from a map[string]string
func selectRandomFromMapOfStrings(m map[string]string) string {
	mapLength := len(m)
	allValues := make([]string, 0, mapLength)
	for _, value := range m {
		allValues = append(allValues, value)
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return allValues[r1.Intn(mapLength)]
}

func isSelectRandom(opts ...bool) bool {
	selectRandomFlag := false
	for i, val := range opts {
		switch i {
		case 0:
			selectRandomFlag = val
		default:
			fmt.Printf("Option index is not defined. Index: %v, value: %v\n", i, val)
		}
	}
	return selectRandomFlag
}

type Identifiable interface {
	Id() string
}

const RepositoriesEndpoint = "artifactory/api/repositories/"

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.Head(RepositoriesEndpoint + id)
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
