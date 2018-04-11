package artifactory

import (
	"fmt"
	"strings"
)

func validateLowerCase(value interface{}, key string) (ws []string, es []error) {
	m := value.(string)
	low := strings.ToLower(m)

	if m != low {
		es = append(es, fmt.Errorf("%s should be lowercase", key))
	}
	return
}
