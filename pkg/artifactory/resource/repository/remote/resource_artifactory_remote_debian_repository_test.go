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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteDebianRepository(t *testing.T) {
	// Exercise the curation attributes through the full create/update/import
	// lifecycle. curated=true with pass_through=true complements the shared
	// curation_test.go loop (which covers true/false) and confirms pass_through
	// is managed independently without drift.
	resource.Test(mkNewRemoteTestCase(repository.DebianPackageType, t, map[string]interface{}{
		"curated":      true,
		"pass_through": true,
	}))
}

func TestAccRemoteDebianRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-debian-remote", "artifactory_remote_debian_repository")

	// Note: use a Debian mirror that Artifactory's curation service does NOT
	// auto-enroll. Known-curatable upstreams (e.g. archive.ubuntu.com,
	// deb.debian.org) are force-enabled to curated=true server-side, which would
	// conflict with the schema default of false and produce a spurious plan diff
	// on upgrade.
	const temp = `
		resource "artifactory_remote_debian_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "http://ftp.debian.org/debian/"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteDebianRepository_migrate_from_SDKv2", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.3",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "http://ftp.debian.org/debian/"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
