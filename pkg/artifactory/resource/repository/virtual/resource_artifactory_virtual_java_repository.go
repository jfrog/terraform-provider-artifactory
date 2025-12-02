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

package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type CommonJavaVirtualRepositoryParams struct {
	ForceMavenAuthentication             bool   `json:"forceMavenAuthentication,omitempty"`
	PomRepositoryReferencesCleanupPolicy string `hcl:"pom_repository_references_cleanup_policy" json:"pomRepositoryReferencesCleanupPolicy,omitempty"`
	KeyPair                              string `hcl:"key_pair" json:"keyPair,omitempty"`
}

type JavaVirtualRepositoryParams struct {
	RepositoryBaseParams
	CommonJavaVirtualRepositoryParams
}

var JavaSchema = map[string]*schema.Schema{
	"force_maven_authentication": {
		Type:        schema.TypeBool,
		Computed:    true,
		Optional:    true,
		Description: "User authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests.",
	},
	"pom_repository_references_cleanup_policy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.StringInSlice(
			[]string{"discard_active_reference", "discard_any_reference", "nothing"}, false,
		),
		Description: "(1: discard_active_reference) Discard Active References - Removes repository elements that are declared directly under project or under a profile in the same POM that is activeByDefault.\n" +
			"(2: discard_any_reference) Discard Any References - Removes all repository elements regardless of whether they are included in an active profile or not.\n" +
			"(3: nothing) Nothing - Does not remove any repository elements declared in the POM.",
	},
	"key_pair": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The keypair used to sign artifacts",
	},
}

func ResourceArtifactoryVirtualJavaRepository(packageType string) *schema.Resource {
	var mavenSchema = lo.Assign(
		JavaSchema,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	var mavenSchemas = GetSchemas(mavenSchema)

	var unpackMavenVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := JavaVirtualRepositoryParams{
			RepositoryBaseParams: UnpackBaseVirtRepo(s, packageType),
			CommonJavaVirtualRepositoryParams: CommonJavaVirtualRepositoryParams{
				KeyPair:                              d.GetString("key_pair", false),
				ForceMavenAuthentication:             d.GetBool("force_maven_authentication", false),
				PomRepositoryReferencesCleanupPolicy: d.GetString("pom_repository_references_cleanup_policy", false),
			},
		}
		repo.PackageType = packageType

		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &JavaVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: packageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		mavenSchemas,
		packer.Default(mavenSchemas[CurrentSchemaVersion]),
		unpackMavenVirtualRepository,
		constructor,
	)
}
