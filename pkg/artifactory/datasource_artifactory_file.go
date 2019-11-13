package artifactory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/schema"
	"io"
	"os"
)

func dataSourceArtifactoryFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArtifactRead,

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
			"output_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"force_overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func dataSourceArtifactRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repository := d.Get("repository").(string)
	path := d.Get("path").(string)
	outputPath := d.Get("output_path").(string)
	forceOverwrite := d.Get("force_overwrite").(bool)

	fileInfo, _, err := c.V1.Artifacts.FileInfo(context.Background(), repository, path)
	if err != nil {
		return err
	}

	skip, err := SkipDownload(fileInfo, outputPath)
	if err != nil && !forceOverwrite {
		return err
	}

	if !skip {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		defer outFile.Close()

		fileInfo, _, err = c.V1.Artifacts.FileContents(context.Background(), repository, path, outFile)
		if err != nil {
			return err
		}
	}

	return packFileInfo(fileInfo, d)
}

func SkipDownload(fileInfo *v1.FileInfo, path string) (bool, error) {
	const skip = true
	const dontSkip = false

	if path == "" {
		// no path specified, nothing to download
		return skip, nil
	}

	if FileExists(path) {
		chks_matches, err := VerifySha256Checksum(path, *fileInfo.Checksums.Sha256)

		if chks_matches {
			return skip, nil
		} else if err != nil {
			return dontSkip, err
		} else {
			return dontSkip, fmt.Errorf("Local file differs from upstream version")
		}
	} else {
		return dontSkip, nil
	}
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func VerifySha256Checksum(path string, expectedSha256 string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	return hex.EncodeToString(hasher.Sum(nil)) == expectedSha256, nil
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
