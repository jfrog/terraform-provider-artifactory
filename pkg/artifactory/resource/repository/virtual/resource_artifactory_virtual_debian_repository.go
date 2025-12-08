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
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var debianSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	repository.PrimaryKeyPairRefSDKv2,
	repository.SecondaryKeyPairRefSDKv2,
	map[string]*schema.Schema{
		"optional_index_compression_formats": {
			Type:     schema.TypeSet,
			Optional: true,
			MinItems: 0,
			Computed: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"bz2", "lzma", "xz"}, false),
			},
			Description: `Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are 'bz2','lzma' and 'xz'. Default value is 'bz2'.`,
		},
		"debian_default_architectures": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "amd64,i386",
			ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringIsNotEmpty, validation.StringMatch(regexp.MustCompile(`.+(?:,.+)*`), "must be comma separated string"))),
			StateFunc:        utilsdk.FormatCommaSeparatedString,
			Description:      `Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.`,
		},
	}, repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DebianPackageType),
)

var DebianSchemas = GetSchemas(debianSchema)

func ResourceArtifactoryVirtualDebianRepository() *schema.Resource {

	type DebianVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		repository.PrimaryKeyPairRefParam
		repository.SecondaryKeyPairRefParam
		OptionalIndexCompressionFormats []string `json:"optionalIndexCompressionFormats"`
		DebianDefaultArchitectures      string   `json:"debianDefaultArchitectures"`
	}

	var unpackDebianVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, repository.DebianPackageType),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
			},
			SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
				SecondaryKeyPairRefSDKv2: d.GetString("secondary_keypair_ref", false),
			},
			OptionalIndexCompressionFormats: d.GetSet("optional_index_compression_formats"),
			DebianDefaultArchitectures:      d.GetString("debian_default_architectures", false),
		}
		repo.PackageType = repository.DebianPackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.DebianPackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		DebianSchemas,
		packer.Default(DebianSchemas[CurrentSchemaVersion]),
		unpackDebianVirtualRepository,
		constructor,
	)
}
