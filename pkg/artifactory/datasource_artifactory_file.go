package artifactory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FileInfo struct {
	Repo              string    `json:"repo,omitempty"`
	Path              string    `json:"path,omitempty"`
	Created           string    `json:"created,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	LastModified      string    `json:"lastModified,omitempty"`
	ModifiedBy        string    `json:"modifiedBy,omitempty"`
	LastUpdated       string    `json:"lastUpdated,omitempty"`
	DownloadUri       string    `json:"downloadUri,omitempty"`
	MimeType          string    `json:"mimeType,omitempty"`
	Size              int       `json:"size,string,omitempty"`
	Checksums         Checksums `json:"checksums,omitempty"`
	OriginalChecksums Checksums `json:"originalChecksums,omitempty"`
	Uri               string    `json:"uri,omitempty"`
}
type Checksums struct {
	Md5    string `json:"md5,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

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
	repository := d.Get("repository").(string)
	path := d.Get("path").(string)
	outputPath := d.Get("output_path").(string)
	forceOverwrite := d.Get("force_overwrite").(bool)
	fileInfo := FileInfo{}
	_, err := m.(*resty.Client).R().SetResult(&fileInfo).Get(fmt.Sprintf("artifactory/api/storage/%s/%s", repository, path))
	if err != nil {
		return err
	}

	fileExists := FileExists(outputPath)
	chksMatches, _ := VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)

	if fileExists {
		if !chksMatches && !forceOverwrite {
			return fmt.Errorf("local file differs from upstream version and no overwrite is permitted")
		}
	} else {
		outdir := filepath.Dir(outputPath)
		err = os.MkdirAll(outdir, os.ModePerm)
		if err != nil {
			return err
		}
		outFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		defer func(outFile *os.File) {
			_ = outFile.Close()
		}(outFile)
	}

	_, err = m.(*resty.Client).R().SetOutput(outputPath).Get(fileInfo.DownloadUri)
	if err != nil {
		return err
	}
	chksMatches, _ = VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
	if !chksMatches {
		return fmt.Errorf("%s checksum and %s checksum do not match, expectd %s", outputPath,fileInfo.DownloadUri, fileInfo.Checksums.Sha256)
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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	return hex.EncodeToString(hasher.Sum(nil)) == expectedSha256, nil
}
