package replication_test

import (
	"github.com/go-resty/resty/v2"

	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/replication"
)

func repConfigExists(id string, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().Head(replication.ReplicationEndpointPath + id)
	return err == nil, err
}
