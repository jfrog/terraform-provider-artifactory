package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/template"
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

func MergeMaps(schemata ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

func CopyInterfaceMap(source map[string]interface{}, target map[string]interface{}) map[string]interface{} {
	for k, v := range source {
		target[k] = v
	}
	return target
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

func ExecuteTemplate(name, temp string, fields interface{}) string {
	var tpl bytes.Buffer
	if err := template.Must(template.New(name).Parse(temp)).Execute(&tpl, fields); err != nil {
		panic(err)
	}

	return tpl.String()
}

func MkNames(name, resource string) (int, string, string) {
	id := RandomInt()
	n := fmt.Sprintf("%s%d", name, id)
	return id, fmt.Sprintf("%s.%s", resource, n), n
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

func BoolPtr(v bool) *bool { return &v }

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

const RepositoriesEndpoint = "artifactory/api/repositories/"

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.Head(RepositoriesEndpoint + id)
}
