package artifactory

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceData struct{ *schema.ResourceData }

func (d *ResourceData) GetStringRef(key string) *string {
	if v, ok := d.GetOkExists(key); ok {
		return artifactory.String(v.(string))
	}
	return nil
}

func (d *ResourceData) GetBoolRef(key string) *bool {
	if v, ok := d.GetOkExists(key); ok {
		return artifactory.Bool(v.(bool))
	}
	return nil
}

func (d *ResourceData) GetIntRef(key string) *int {
	if v, ok := d.GetOkExists(key); ok {
		return artifactory.Int(v.(int))
	}
	return nil
}

func (d *ResourceData) GetSetRef(key string) *[]string {
	if v, ok := d.GetOkExists(key); ok {
		arr := CastToStringArr(v.(*schema.Set).List())
		return &arr
	}
	return new([]string)
}

func (d *ResourceData) GetListRef(key string) *[]string {
	if v, ok := d.GetOkExists(key); ok {
		arr := CastToStringArr(v.([]interface{}))
		return &arr
	}
	return new([]string)
}

func (d *ResourceData) SetOrPropagate(err *error) func(string, interface{}) {
	return func(key string, v interface{}) {
		if *err != nil {
			return
		}

		*err = d.Set(key, v)
	}
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

func GetMD5Hash(o interface{}) string {
	if len(o.(string)) == 0 { // Don't hash empty strings
		return ""
	}

	hasher := sha256.New()
	hasher.Write([]byte(o.(string)))
	hasher.Write([]byte("OQ9@#9i4$c8g$4^n%PKT8hUva3CC^5"))
	return hex.EncodeToString(hasher.Sum(nil))
}
