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
	"github.com/samber/lo"
)

var bowerSchema = lo.Assign(
	externalDependenciesSchema,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.BowerPackageType),
)

var BowerSchemas = GetSchemas(bowerSchema)

func ResourceArtifactoryVirtualBowerRepository() *schema.Resource {
	var unpackBowerVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := unpackExternalDependenciesVirtualRepository(s, repository.BowerPackageType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &ExternalDependenciesVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.BowerPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		BowerSchemas,
		packer.Default(BowerSchemas[CurrentSchemaVersion]),
		unpackBowerVirtualRepository,
		constructor,
	)
}
