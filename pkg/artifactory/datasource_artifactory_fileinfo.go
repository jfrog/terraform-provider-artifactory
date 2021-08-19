package artifactory

import (
	"context"
	"fmt"

	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceArtifactoryFileInfo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileInfoRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceFileInfoRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repository := d.Get("repository").(string)
	path := d.Get("path").(string)

	fileInfo, _, err := c.V1.Artifacts.FileInfo(context.Background(), repository, path)
	if err != nil {
		return err
	}

	return packFileInfo(fileInfo, d)
}

func packFileInfo(fileInfo *v1.FileInfo, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	d.SetId(*fileInfo.DownloadUri)

	logErr(d.Set("created", *fileInfo.Created))
	logErr(d.Set("created_by", *fileInfo.CreatedBy))
	logErr(d.Set("last_modified", *fileInfo.LastModified))
	logErr(d.Set("modified_by", *fileInfo.ModifiedBy))
	logErr(d.Set("last_updated", *fileInfo.LastUpdated))
	logErr(d.Set("download_uri", *fileInfo.DownloadUri))
	logErr(d.Set("mimetype", *fileInfo.MimeType))
	logErr(d.Set("size", *fileInfo.Size))

	if fileInfo.Checksums != nil {
		logErr(d.Set("md5", *fileInfo.Checksums.Md5))
		logErr(d.Set("sha1", *fileInfo.Checksums.Sha1))
		logErr(d.Set("sha256", *fileInfo.Checksums.Sha256))
	}

	if hasErr {
		return fmt.Errorf("failed to pack fileInfo")
	}

	return nil
}
