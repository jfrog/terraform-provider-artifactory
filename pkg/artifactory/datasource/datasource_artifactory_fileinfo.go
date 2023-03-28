package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ArtifactoryFileInfo() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFileInfoRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: repository.RepoKeyValidator,
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
	_, err := m.(util.ProvderMetadata).Client.R().
		SetResult(&fileInfo).
		SetPathParams(map[string]string{
			"repoKey": repo,
			"path":    path,
		}).
		Get("artifactory/api/storage/{repoKey}/{path}")
	if err != nil {
		return diag.FromErr(err)
	}

	return packFileInfo(fileInfo, d)
}

func packFileInfo(fileInfo FileInfo, d *schema.ResourceData) diag.Diagnostics {
	setValue := util.MkLens(d)

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

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack fileInfo %q", errors)
	}

	return nil
}
