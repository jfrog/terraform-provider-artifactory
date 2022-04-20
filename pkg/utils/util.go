package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceData struct{ *schema.ResourceData }

func (d *ResourceData) GetString(key string, onlyIfChanged bool) string {
	if v, ok := d.GetOk(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(string)
	}
	return ""
}

func BoolPtr(v bool) *bool { return &v }

func (d *ResourceData) GetBoolRef(key string, onlyIfChanged bool) *bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return BoolPtr(v.(bool))
	}
	return nil
}

func (d *ResourceData) GetBool(key string, onlyIfChanged bool) bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(bool)
	}
	return false
}

func (d *ResourceData) GetInt(key string, onlyIfChanged bool) int {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(int)
	}
	return 0
}

func (d *ResourceData) GetSet(key string) []string {
	if v, ok := d.GetOkExists(key); ok {
		arr := CastToStringArr(v.(*schema.Set).List())
		return arr
	}
	return nil
}
func (d *ResourceData) GetList(key string) []string {
	if v, ok := d.GetOkExists(key); ok {
		arr := CastToStringArr(v.([]interface{}))
		return arr
	}
	return []string{}
}

func CastToStringArr(arr []interface{}) []string {
	cpy := make([]string, 0, len(arr))
	for _, r := range arr {
		cpy = append(cpy, r.(string))
	}

	return cpy
}

func CastToInterfaceArr(arr []string) []interface{} {
	cpy := make([]interface{}, 0, len(arr))
	for _, r := range arr {
		cpy = append(cpy, r)
	}

	return cpy
}

func RandomInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10000000)
}

func RandBool() bool {
	return RandomInt()%2 == 0
}

func RandSelect(items ...interface{}) interface{} {
	return items[RandomInt()%len(items)]
}

func MergeSchema(schemata ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := map[string]*schema.Schema{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

type Lens func(key string, value interface{}) []error

type Schema map[string]*schema.Schema

func SchemaHasKey(skeema map[string]*schema.Schema) HclPredicate {
	return func(key string) bool {
		_, ok := skeema[key]
		return ok
	}
}

type HclPredicate func(hcl string) bool

func MkLens(d *schema.ResourceData) Lens {
	var errors []error
	return func(key string, value interface{}) []error {
		if err := d.Set(key, value); err != nil {
			errors = append(errors, err)
		}
		return errors
	}
}

func SendConfigurationPatch(content []byte, m interface{}) error {
	_, err := m.(*resty.Client).R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		Patch("artifactory/api/system/configuration")

	return err
}

func FormatCommaSeparatedString(thing interface{}) string {
	fields := strings.Fields(thing.(string))
	sort.Strings(fields)
	return strings.Join(fields, ",")
}

func BuildResty(URL, version string) (*resty.Client, error) {
	u, err := url.ParseRequestURI(URL)

	if err != nil {
		return nil, err
	}

	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	restyBase := resty.New().SetHostURL(baseUrl).OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		if response == nil {
			return fmt.Errorf("no response found")
		}

		if response.StatusCode() >= http.StatusBadRequest {
			return fmt.Errorf("\n%d %s %s\n%s", response.StatusCode(), response.Request.Method, response.Request.URL, string(response.Body()[:]))
		}
		return nil
	}).
		SetHeader("content-type", "application/json").
		SetHeader("accept", "*/*").
		SetHeader("user-agent", "jfrog/terraform-provider-artifactory:"+version).
		SetRetryCount(5)

	restyBase.DisableWarn = true

	return restyBase, nil
}

func AddAuthToResty(client *resty.Client, apiKey, accessToken string) (*resty.Client, error) {
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
	}
	if apiKey != "" {
		return client.SetHeader("X-JFrog-Art-Api", apiKey), nil
	}
	return nil, fmt.Errorf("no authentication details supplied")
}

var NeverRetry = func(response *resty.Response, err error) bool {
	return false
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func VerifySha256Checksum(path string, expectedSha256 string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	return hex.EncodeToString(hasher.Sum(nil)) == expectedSha256, nil
}

type SupportedRepoClasses struct {
	RepoLayoutRef      string
	SupportedRepoTypes map[string]bool
}

//Consolidated list of Default Repo Layout for all Package Types with active Repo Types
var defaultRepoLayoutMap = map[string]SupportedRepoClasses{
	"alpine":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"bower":     {RepoLayoutRef: "bower-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cran":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cargo":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"chef":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cocoapods": {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"composer":  {RepoLayoutRef: "composer-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conan":     {RepoLayoutRef: "conan-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conda":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"debian":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"docker":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gems":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"generic":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gitlfs":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"go":        {RepoLayoutRef: "go-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gradle":    {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"helm":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"ivy":       {RepoLayoutRef: "ivy-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"maven":     {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"npm":       {RepoLayoutRef: "npm-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"nuget":     {RepoLayoutRef: "nuget-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"opkg":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"p2":        {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true, "virtual": true}},
	"pub":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"puppet":    {RepoLayoutRef: "puppet-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"pypi":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"sbt":       {RepoLayoutRef: "sbt-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"vagrant":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "federated": true}},
	"vcs":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true}},
	"rpm":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
}

//Return the default repo layout by Repository Type & Package Type
func GetDefaultRepoLayoutRef(repositoryType string, packageType string) func() (interface{}, error) {
	return func() (interface{}, error) {
		if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
			return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
		}
		return "", fmt.Errorf("default repo layout not found for repository type %v & package type %v", repositoryType, packageType)
	}
}

const (
	KeypairEndPoint = "artifactory/api/security/keypair/"
)

func VerifyKeyPair(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(KeypairEndPoint + id)
}

var GradleLikeRepoTypes = []string{
	"gradle",
	"sbt",
	"ivy",
}
