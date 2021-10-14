package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"math"
	"reflect"
	"strings"
)

func fmtMapToHcl(fields map[string]interface{}) string {
	var allPairs []string
	max := float64(0)
	for key, _ := range fields {
		max = math.Max(max, float64(len(key)))
	}
	for key, value := range fields {
		hcl := toHclFormat(value)
		format := toHclFormatString(3, int(max), value)
		allPairs = append(allPairs, fmt.Sprintf(format, key, hcl))
	}

	return strings.Join(allPairs, "\n")
}
func toHclFormatString(tabs, max int, value interface{}) string {
	prefix := ""
	suffix := ""
	delimeter := "="
	if reflect.TypeOf(value).Kind() == reflect.Map {
		delimeter = ""
		prefix = "{"
		suffix = "}"
	}
	return fmt.Sprintf("%s%%-%ds %s %s%s%s", strings.Repeat("\t", tabs), max, delimeter, prefix, "%s", suffix)
}
func mapToTestChecks(fqrn string, fields map[string]interface{}) []resource.TestCheckFunc {
	var result []resource.TestCheckFunc
	for key, value := range fields {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			for i, lv := range value.([]interface{}) {
				result = append(result, resource.TestCheckResourceAttr(
					fqrn,
					fmt.Sprintf("%s.%d", key, i),
					fmt.Sprintf("%v", lv),
				))
			}
		case reflect.Map:
			// this also gets generated, but it's value is '1', which is also the size. So, I don't know
			// what it means
			// content_synchronisation.0.%
			resource.TestCheckResourceAttr(
				fqrn,
				fmt.Sprintf("%s.#", key),
				fmt.Sprintf("%d", len(value.(map[string]interface{}))),
			)
		default:
			result = append(result, resource.TestCheckResourceAttr(fqrn, key, fmt.Sprintf(`%v`, value)))
		}
	}
	return result
}
func toHclFormat(thing interface{}) string {
	switch thing.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, thing.(string))
	case []interface{}:
		var result []string
		for _, e := range thing.([]interface{}) {
			result = append(result, toHclFormat(e))
		}
		return fmt.Sprintf("[%s]", strings.Join(result, ","))
	case map[string]interface{}:
		return fmt.Sprintf("\n\t%s\n\t\t\t\t", fmtMapToHcl(thing.(map[string]interface{})))
	default:
		return fmt.Sprintf("%v", thing)
	}
}