package artifactory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"io"
	"os"
)

func dataSourceArtifactoryFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileRead,

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
				Required: true,
			},
			"force_overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func dataSourceFileRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repository := d.Get("repository").(string)
	path := d.Get("path").(string)
	outputPath := d.Get("output_path").(string)
	forceOverwrite := d.Get("force_overwrite").(bool)

	fileInfo, _, err := c.V1.Artifacts.FileInfo(context.Background(), repository, path)
	if err != nil {
		return err
	}

	fileExists := FileExists(outputPath)
	chksMatches, _ := VerifySha256Checksum(outputPath, *fileInfo.Checksums.Sha256)

	if !fileExists || (!chksMatches && forceOverwrite) {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		defer outFile.Close()

		fileInfo, _, err = c.V1.Artifacts.FileContents(context.Background(), repository, path, outFile)
		if err != nil {
			return err
		}
	} else if !chksMatches {
		return fmt.Errorf("Local file differs from upstream version")
	}

	return packFileInfo(fileInfo, d)
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
