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

var goSchema = lo.Assign(
	map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Default:     true,
			Optional:    true,
			Description: "When set (default), Artifactory will automatically follow remote VCS roots in 'go-import' meta tags to download remote modules.",
		},
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.GoPackageType),
)

var GoSchemas = GetSchemas(goSchema)

func ResourceArtifactoryVirtualGoRepository() *schema.Resource {
	type GoVirtualRepositoryParams struct {
		RepositoryBaseParams
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled,omitempty"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns,omitempty"`
	}

	var unpackGoVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := GoVirtualRepositoryParams{
			RepositoryBaseParams:         UnpackBaseVirtRepo(s, repository.GoPackageType),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
		}
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &GoVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.GoPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		GoSchemas,
		packer.Default(GoSchemas[CurrentSchemaVersion]),
		unpackGoVirtualRepository,
		constructor,
	)
}
