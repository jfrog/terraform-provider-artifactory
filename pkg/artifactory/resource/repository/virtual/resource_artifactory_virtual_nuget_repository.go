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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var nugetSchema = lo.Assign(
	map[string]*schema.Schema{
		"force_nuget_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If set, user authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.NugetPackageType),
)

var NugetSchemas = GetSchemas(nugetSchema)

func ResourceArtifactoryVirtualNugetRepository() *schema.Resource {

	type NugetVirtualRepositoryParams struct {
		RepositoryBaseParams
		ForceNugetAuthentication bool `json:"forceNugetAuthentication"`
	}

	var unpackNugetVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := NugetVirtualRepositoryParams{
			RepositoryBaseParams:     UnpackBaseVirtRepo(s, repository.NugetPackageType),
			ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
		}
		repo.PackageType = repository.NugetPackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NugetVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.NugetPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		NugetSchemas,
		packer.Default(NugetSchemas[CurrentSchemaVersion]),
		unpackNugetVirtualRepository,
		constructor,
	)
}
