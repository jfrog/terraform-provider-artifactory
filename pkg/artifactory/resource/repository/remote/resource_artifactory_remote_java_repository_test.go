package remote_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
)

func TestAccRemoteGradleLikeRepository(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
				"missed_cache_period_seconds":     1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
				"metadata_retrieval_timeout_secs": 30,   // https://github.com/jfrog/terraform-provider-artifactory/issues/509
				"list_remote_folder_items":        true,
				"max_unique_snapshots":            6,
			}))
		})
	}
}
