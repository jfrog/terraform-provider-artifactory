// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
)

func TestAccRemoteMavenRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.MavenPackageType, t, map[string]interface{}{
		"missed_cache_period_seconds":     1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"metadata_retrieval_timeout_secs": 30,   // https://github.com/jfrog/terraform-provider-artifactory/issues/509
		"list_remote_folder_items":        true,
		"max_unique_snapshots":            6,
		"curated":                         false,
		"pass_through":                    false,
	}))
}
