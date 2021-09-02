package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceArtifactoryFileInfo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileInfoRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: repoKeyValidator,
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
	c := m.(*ArtClient).Resty

	repository := d.Get("repository").(string)
	path := d.Get("path").(string)

	fileInfo := FileInfo{}
	_, err := c.R().SetResult(&fileInfo).Get(fmt.Sprintf("artifactory/api/storage/%s/%s", repository, path))
	if err != nil {
		return err
	}

	return packFileInfo(fileInfo, d)
}

func packFileInfo(fileInfo FileInfo, d *schema.ResourceData) error {
	setValue := mkLens(d)

	d.SetId(fileInfo.DownloadUri)

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
		return fmt.Errorf("failed to pack fileInfo %q", errors)
	}

	return nil
}
