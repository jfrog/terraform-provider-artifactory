package artifactory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

func (fi FileInfo) Id() string {
	return fi.Repo + fi.Path
}

func dataSourceArtifactoryFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFileReader,
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the repository where the file is stored.",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to the file within the repository.",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time & date when the file was created.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who created the file.",
			},
			"last_modified": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time & date when the file was last modified.",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who last modified the file.",
			},
			"last_updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time & date when the file was last updated.",
			},
			"download_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URI that can be used to download the file.",
			},
			"mimetype": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The mimetype of the file.",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the file.",
			},
			"md5": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MD5 checksum of the file.",
			},
			"sha1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA1 checksum of the file.",
			},
			"sha256": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA256 checksum of the file.",
			},
			"output_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The local path the file should be downloaded to.",
			},
			"force_overwrite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, an existing file in the output_path will be overwritten.",
			},
			"path_is_aliased": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "If set to `true`, the provider will get the artifact path directly from Artifactory without verification " +
					"or de aliasing. More details in the [official documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-RetrieveLatestArtifact)",
			},
		},
	}
}

func dataSourceFileReader(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	repository := d.Get("repository").(string)
	path := d.Get("path").(string)
	outputPath := d.Get("output_path").(string)
	forceOverwrite := d.Get("force_overwrite").(bool)
	pathIsAliased := d.Get("path_is_aliased").(bool)
	fileInfo := FileInfo{}

	if !pathIsAliased {
		_, err := m.(*resty.Client).R().SetResult(&fileInfo).Get(fmt.Sprintf("artifactory/api/storage/%s/%s", repository, path))
		if err != nil {
			return diag.FromErr(err)
		}

		fileExists := FileExists(outputPath)
		chksMatches, _ := VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)

		/*--File Download logic--
		1. File doesn't exist
		2. In Data Source argument `force_overwrite` set to true, an existing file in the output_path will be overwritten. Ignore file exists or not
		3. File exists but check sum doesn't match
		*/
		if !fileExists || forceOverwrite || (fileExists && !chksMatches) {
			outdir := filepath.Dir(outputPath)
			err = os.MkdirAll(outdir, os.ModePerm)
			if err != nil {
				return diag.FromErr(err)
			}
			outFile, err := os.Create(outputPath)
			if err != nil {
				return diag.FromErr(err)
			}
			defer func(outFile *os.File) {
				_ = outFile.Close()
			}(outFile)
		} else { //download not required
			d.SetId(fileInfo.Id())
			return diag.Diagnostics{{
				Severity: diag.Warning,
				Summary:  "WARN-001: file download skipped.",
				Detail:   fmt.Sprintf("WARN-001: file download skipped. fileExists: %v, chksMatches: %v, forceOverwrite: %v", fileExists, chksMatches, forceOverwrite),
			}}

		}
		_, err = m.(*resty.Client).R().SetOutput(outputPath).Get(fileInfo.DownloadUri)
		if err != nil {
			return diag.FromErr(err)
		}

		chksMatches, _ = VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
		if !chksMatches {
			return diag.FromErr(fmt.Errorf("%s checksum and %s checksum do not match, expectd %s", outputPath, fileInfo.DownloadUri, fileInfo.Checksums.Sha256))
		}
	} else { // if we download the latest artifact (use path_is_aliased), we don't have all the data for the fileInfo struct, because no GET call was sent.
		fileInfo.Repo = repository
		fileInfo.Path = path
		d.SetId(fileInfo.Path)
		fileExists := FileExists(outputPath)

		if !fileExists || forceOverwrite {
			outdir := filepath.Dir(outputPath)
			err := os.MkdirAll(outdir, os.ModePerm)
			if err != nil {
				return diag.FromErr(err)
			}
			outFile, err := os.Create(outputPath)
			if err != nil {
				return diag.FromErr(err)
			}
			defer func(outFile *os.File) {
				_ = outFile.Close()
			}(outFile)
		} else { //download not required
			d.SetId(fileInfo.Path)
			return diag.Diagnostics{{
				Severity: diag.Warning,
				Summary:  "WARN-001: file download skipped.",
				Detail:   fmt.Sprintf("WARN-001: file download skipped. fileExists: %v, forceOverwrite: %v", fileExists, forceOverwrite),
			}}

		}
		_, err := m.(*resty.Client).R().SetOutput(outputPath).Get(fmt.Sprintf("artifactory/%s/%s", repository, path))
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	return diag.FromErr(packFileInfo(fileInfo, d))
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
