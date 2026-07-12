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

package artifact

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func ArtifactoryFileInfo() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFileInfoRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validator.RepoKey,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"download_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mimetype": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"md5": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFileInfoRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	repo := d.Get("repository").(string)
	path := d.Get("path").(string)

	fileInfo := FileInfo{}
	resp, err := m.(util.ProviderMetadata).Client.R().
		SetResult(&fileInfo).
		SetPathParams(map[string]string{
			"repoKey": repo,
			"path":    path,
		}).
		Get("artifactory/api/storage/{repoKey}/{path}")
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.IsError() {
		return diag.Errorf("%s", resp.String())
	}

	return packFileInfo(fileInfo, d)
}

func packFileInfo(fileInfo FileInfo, d *schema.ResourceData) diag.Diagnostics {
	setValue := utilsdk.MkLens(d)

	d.SetId(fileInfo.Id())

	setValue("created", fileInfo.Created)
	setValue("created_by", fileInfo.CreatedBy)
	setValue("last_modified", fileInfo.LastModified)
	setValue("modified_by", fileInfo.ModifiedBy)
	setValue("last_updated", fileInfo.LastUpdated)
	setValue("download_uri", fileInfo.DownloadUri)
	setValue("mimetype", fileInfo.MimeType)
	errors := setValue("size", fileInfo.Size)

	if fileInfo.Checksums.Md5 != "" {
		errors = setValue("md5", fileInfo.Checksums.Md5)
	}
	if fileInfo.Checksums.Sha1 != "" {
		errors = setValue("sha1", fileInfo.Checksums.Sha1)
	}
	if fileInfo.Checksums.Sha256 != "" {
		errors = setValue("sha256", fileInfo.Checksums.Sha256)
	}

	if len(errors) > 0 {
		return diag.Errorf("failed to pack fileInfo %q", errors)
	}

	return nil
}
