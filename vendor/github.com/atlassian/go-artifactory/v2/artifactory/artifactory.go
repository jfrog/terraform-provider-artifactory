package artifactory

import (
	"github.com/atlassian/go-artifactory/v2/artifactory/client"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/atlassian/go-artifactory/v2/artifactory/v2"

	"net/http"
)

// Artifactory is the container for all the api methods
type Artifactory struct {
	V1 *v1.V1
	V2 *v2.V2
}

// NewClient creates a Artifactory from a provided base url for an artifactory instance and a service Artifactory
func NewClient(baseURL string, httpClient *http.Client) (*Artifactory, error) {
	c, err := client.NewClient(baseURL, httpClient)

	if err != nil {
		return nil, err
	}

	rt := &Artifactory{
		V1: v1.NewV1(c),
		V2: v2.NewV2(c),
	}

	return rt, nil
}
