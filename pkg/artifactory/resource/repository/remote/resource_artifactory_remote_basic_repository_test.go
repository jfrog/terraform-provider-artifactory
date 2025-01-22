package remote_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
)

func TestAccRemoteLikeBasicRepository(t *testing.T) {
	for _, repoType := range remote.PackageTypesLikeBasic {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(mkNewRemoteTestCase(repoType, t, map[string]interface{}{
				"missed_cache_period_seconds": 1800,
			}))
		})
	}
}
