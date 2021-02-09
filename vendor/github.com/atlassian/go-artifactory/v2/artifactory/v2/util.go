package v2

import "github.com/atlassian/go-artifactory/v2/artifactory/client"

func String(v string) *string { return &v }

func NewV2(client *client.Client) *V2 {
	v := &V2{}
	v.common.client = client

	v.Security = (*SecurityService)(&v.common)

	return v
}
