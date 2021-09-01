package artifactory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"math/rand"
	"text/template"
	"time"
)

type ResourceData struct{ *schema.ResourceData }

func (d *ResourceData) getStringRef(key string, onlyIfChanged bool) *string {
	if v, ok := d.GetOk(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.String(v.(string))
	}
	return nil
}
func (d *ResourceData) getString(key string, onlyIfChanged bool) string {
	if v, ok := d.GetOk(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(string)
	}
	return ""
}

func (d *ResourceData) getBoolRef(key string, onlyIfChanged bool) *bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.Bool(v.(bool))
	}
	return nil
}

func (d *ResourceData) getBool(key string, onlyIfChanged bool) bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(bool)
	}
	return false
}

func (d *ResourceData) getIntRef(key string, onlyIfChanged bool) *int {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.Int(v.(int))
	}
	return nil
}

func (d *ResourceData) getInt(key string, onlyIfChanged bool) int {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return v.(int)
	}
	return 0
}

func (d *ResourceData) getSetRef(key string) *[]string {
	if v, ok := d.GetOkExists(key); ok {
		arr := castToStringArr(v.(*schema.Set).List())
		return &arr
	}
	return new([]string)
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
func (d *ResourceData) getListRef(key string) *[]string {
	if v, ok := d.GetOkExists(key); ok {
		arr := castToStringArr(v.([]interface{}))
		return &arr
	}
	return new([]string)
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

func getMD5Hash(o interface{}) string {
	if len(o.(string)) == 0 { // Don't hash empty strings
		return ""
	}

	hasher := sha256.New()
	hasher.Write([]byte(o.(string)))
	hasher.Write([]byte("OQ9@#9i4$c8g$4^n%PKT8hUva3CC^5"))
	encodeToString := hex.EncodeToString(hasher.Sum(nil))
	return encodeToString
}

var randomInt = func() func() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int
}()

func mergeSchema(schemata ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := map[string]*schema.Schema{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

func repoExists(id string, m interface{}) (bool, error) {
	client := m.(*ArtClient).Resty

	_, err := client.R().Head(repositoriesEndpoint+ id)

	return err == nil, err
}


func executeTemplate(name, temp string, fields interface{} ) string {
	var tpl bytes.Buffer
	if err := template.Must(template.New(name).Parse(temp)).Execute(&tpl,fields); err != nil {
		panic(err)
	}

	return tpl.String()
}

func mkNames(name, resource string) (int, string, string) {
	id := randomInt()
	n := fmt.Sprintf("%s%d", name, id)
	return id, fmt.Sprintf("%s.%s", resource, n), n
}

func mkLens(d *schema.ResourceData) func(key string, value interface{}) []error {
	var errors []error
	return func(key string, value interface{}) []error {
		if err := d.Set(key, value); err != nil {
			errors = append(errors, err)
		}
		return errors
	}
}

func cascadingErr(hasErr *bool) func(error) {
	if hasErr == nil {
		panic("hasError cannot be nil")
	}
	return func(err error) {
		if err != nil {
			fmt.Println(err)
			*hasErr = true
		}
	}
}
