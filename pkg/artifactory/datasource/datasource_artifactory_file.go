package datasource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

func ArtifactoryFile() *schema.Resource {
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
				Description: "If set to `true`, the provider will get the artifact path directly from Artifactory without attempting to resolve " +
					"it or verify it and will delegate this to artifactory if the file exists. More details in the [official documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-RetrieveLatestArtifact)",
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

	tflog.Debug(ctx, "dataSourceFileReader", map[string]interface{}{
		"repository":     repository,
		"path":           path,
		"outputPath":     outputPath,
		"forceOverwrite": forceOverwrite,
		"pathIsAliased":  pathIsAliased,
	})

	if !pathIsAliased {
		tflog.Debug(ctx, "pathIsAliased == false")

		tflog.Debug(ctx, "Fetching file info")
		_, err := m.(*resty.Client).R().SetResult(&fileInfo).Get(fmt.Sprintf("artifactory/api/storage/%s/%s", repository, path))
		if err != nil {
			return diag.FromErr(err)
		}

		fileExists := FileExists(outputPath)
		chksMatches, err := VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Debug(ctx, "File info fetched", map[string]interface{}{
			"fileInfo":    fileInfo,
			"fileExists":  fileExists,
			"chksMatches": chksMatches,
		})

		/*--File Download logic--
		1. File doesn't exist
		2. In Data Source argument `force_overwrite` set to true, an existing file in the output_path will be overwritten. Ignore file exists or not
		3. File exists but check sum doesn't match
		*/
		if !fileExists || forceOverwrite || (fileExists && !chksMatches) {
			tflog.Debug(ctx, "Should download file")
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
			tflog.Debug(ctx, "Skip downloading file")
			d.SetId(fileInfo.Id())
			return diag.Diagnostics{{
				Severity: diag.Warning,
				Summary:  "WARN-001: file download skipped.",
				Detail:   fmt.Sprintf("WARN-001: file download skipped. fileExists: %v, chksMatches: %v, forceOverwrite: %v", fileExists, chksMatches, forceOverwrite),
			}}
		}

		tflog.Debug(ctx, "Downloading file...", map[string]interface{}{
			"fileInfo.DownloadUri": fileInfo.DownloadUri,
			"outputPath":           outputPath,
		})
		_, err = m.(*resty.Client).R().SetOutput(outputPath).Get(fileInfo.DownloadUri)
		if err != nil {
			return diag.FromErr(err)
		}

		chksMatches, err = VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Debug(ctx, "Verify checksum", map[string]interface{}{
			"fileInfo.Checksums.Sha256": fileInfo.Checksums.Sha256,
			"chksMatches":               chksMatches,
		})
		if !chksMatches {
			return diag.Errorf("Checksums for file %s and %s do not match, expected %s", outputPath, fileInfo.DownloadUri, fileInfo.Checksums.Sha256)
		}
	} else { // if we download the latest artifact (use path_is_aliased), we don't have all the data for the fileInfo struct, because no GET call was sent.
		tflog.Debug(ctx, "pathIsAliased == true")

		fileInfo.Repo = repository
		fileInfo.Path = path
		d.SetId(fileInfo.Path)
		fileExists := FileExists(outputPath)

		tflog.Debug(ctx, "File info", map[string]interface{}{
			"fileInfo":   fileInfo,
			"fileExists": fileExists,
		})

		if !fileExists || forceOverwrite {
			tflog.Debug(ctx, "Should download file")
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
			tflog.Debug(ctx, "Skip downloading file")
			d.SetId(fileInfo.Path)
			return diag.Diagnostics{{
				Severity: diag.Warning,
				Summary:  "WARN-001: file download skipped.",
				Detail:   fmt.Sprintf("WARN-001: file download skipped. fileExists: %v, forceOverwrite: %v", fileExists, forceOverwrite),
			}}

		}

		tflog.Debug(ctx, "Downloading file...", map[string]interface{}{
			"repository path": fmt.Sprintf("artifactory/%s/%s", repository, path),
			"outputPath":      outputPath,
		})
		_, err := m.(*resty.Client).R().SetOutput(outputPath).Get(fmt.Sprintf("artifactory/%s/%s", repository, path))
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	tflog.Debug(ctx, "Calling packFileInfo", map[string]interface{}{
		"fileInfo": fileInfo,
	})

	return packFileInfo(fileInfo, d)
}
