package artifactory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ResourceData struct{ *schema.ResourceData }

func (d *ResourceData) getStringRef(key string, onlyIfChanged bool) *string {
	if v, ok := d.GetOk(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.String(v.(string))
	}
	return nil
}

func (d *ResourceData) getBoolRef(key string, onlyIfChanged bool) *bool {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.Bool(v.(bool))
	}
	return nil
}

func (d *ResourceData) getIntRef(key string, onlyIfChanged bool) *int {
	if v, ok := d.GetOkExists(key); ok && (!onlyIfChanged || d.HasChange(key)) {
		return artifactory.Int(v.(int))
	}
	return nil
}

func (d *ResourceData) getSetRef(key string) *[]string {
	if v, ok := d.GetOkExists(key); ok {
		arr := castToStringArr(v.(*schema.Set).List())
		return &arr
	}
	return new([]string)
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
	return hex.EncodeToString(hasher.Sum(nil))
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

func sendConfigurationPatch(content []byte, m interface{}) error {
	c := m.(*ArtClient).ArtNew

	serviceDetails := c.GetConfig().GetServiceDetails()
	httpClientDetails := serviceDetails.CreateHttpClientDetails()
	httpClientDetails.Headers["Content-Type"] = "application/yaml"

	_, _, err := c.Client().SendPatch(fmt.Sprintf("%sapi/system/configuration", serviceDetails.GetUrl()), content, &httpClientDetails)

	return err
}
