package artifactory

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceData struct{ *schema.ResourceData }

func (d *ResourceData) getString(key string, onlyIfChanged bool) string {
	if v, ok := d.GetOk(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(string)
	}
	return ""
}

func (d *ResourceData) getBoolRef(key string, onlyIfChanged bool) *bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return BoolPtr(v.(bool))
	}
	return nil
}

func (d *ResourceData) getBool(key string, onlyIfChanged bool) bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(bool)
	}
	return false
}

func (d *ResourceData) getInt(key string, onlyIfChanged bool) int {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(int)
	}
	return 0
}

func (d *ResourceData) getSet(key string) []string {
	if v, ok := d.GetOkExists(key); ok {
		arr := castToStringArr(v.(*schema.Set).List())
		return arr
	}
	return nil
}
func (d *ResourceData) getList(key string) []string {
	if v, ok := d.GetOkExists(key); ok {
		arr := castToStringArr(v.([]interface{}))
		return arr
	}
	return []string{}
}

func castToStringArr(arr []interface{}) []string {
	cpy := make([]string, 0, len(arr))
	for _, r := range arr {
		cpy = append(cpy, r.(string))
	}

	return cpy
}

func castToInterfaceArr(arr []string) []interface{} {
	cpy := make([]interface{}, 0, len(arr))
	for _, r := range arr {
		cpy = append(cpy, r)
	}

	return cpy
}

func randomInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10000000)
}

func randBool() bool {
	return randomInt()%2 == 0
}

func randSelect(items ...interface{}) interface{} {
	return items[randomInt()%len(items)]
}

func mergeMaps(schemata ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

func copyInterfaceMap(source map[string]interface{}, target map[string]interface{}) map[string]interface{} {
	for k, v := range source {
		target[k] = v
	}
	return target
}

func mergeSchema(schemata ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := map[string]*schema.Schema{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

func executeTemplate(name, temp string, fields interface{}) string {
	var tpl bytes.Buffer
	if err := template.Must(template.New(name).Parse(temp)).Execute(&tpl, fields); err != nil {
		panic(err)
	}

	return tpl.String()
}

func mkNames(name, resource string) (int, string, string) {
	id := randomInt()
	n := fmt.Sprintf("%s%d", name, id)
	return id, fmt.Sprintf("%s.%s", resource, n), n
}

type Lens func(key string, value interface{}) []error

type Schema map[string]*schema.Schema

func schemaHasKey(skeema map[string]*schema.Schema) HclPredicate {
	return func(key string) bool {
		_, ok := skeema[key]
		return ok
	}
}

type HclPredicate func(hcl string) bool

func mkLens(d *schema.ResourceData) Lens {
	var errors []error
	return func(key string, value interface{}) []error {
		if err := d.Set(key, value); err != nil {
			errors = append(errors, err)
		}
		return errors
	}
}

func sendConfigurationPatch(content []byte, m interface{}) error {

	_, err := m.(*resty.Client).R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		Patch("artifactory/api/system/configuration")

	return err
}

func BoolPtr(v bool) *bool { return &v }

func formatCommaSeparatedString(thing interface{}) string {
	fields := strings.Fields(thing.(string))
	sort.Strings(fields)
	return strings.Join(fields, ",")
}
