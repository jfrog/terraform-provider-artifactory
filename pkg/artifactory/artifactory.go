package artifactory

import (
	"fmt"

	"github.com/samber/lo"
)

type artifactoryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e artifactoryError) String() string {
	return fmt.Sprintf("%s - %s", e.Code, e.Message)
}

type ArtifactoryErrorsResponse struct {
	Errors []artifactoryError `json:"errors"`
}

func (r ArtifactoryErrorsResponse) String() string {
	errs := lo.Reduce(r.Errors, func(err string, item artifactoryError, _ int) string {
		if err == "" {
			return item.String()
		} else {
			return fmt.Sprintf("%s, %s", err, item.String())
		}
	}, "")
	return errs
}
