package datasource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
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
	if fi.DownloadUri != "" {
		return fi.DownloadUri
	} else {
		return fi.Repo + fi.Path
	}
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

func createOutputFile(outputPath string) error {
	outdir := filepath.Dir(outputPath)
	err := os.MkdirAll(outdir, os.ModePerm)
	if err != nil {
		return err
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		outFile.Close()
	}(outFile)

	return nil
}

func dataSourceFileReader(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	repository := d.Get("repository").(string)
	path := d.Get("path").(string)
	outputPath := d.Get("output_path").(string)
	forceOverwrite := d.Get("force_overwrite").(bool)
	pathIsAliased := d.Get("path_is_aliased").(bool)

	var fileInfo FileInfo
	var err error

	tflog.Debug(ctx, "dataSourceFileReader", map[string]interface{}{
		"repository":     repository,
		"path":           path,
		"outputPath":     outputPath,
		"forceOverwrite": forceOverwrite,
		"pathIsAliased":  pathIsAliased,
	})

	if !pathIsAliased {
		tflog.Debug(ctx, "pathIsAliased == false")
		fileInfo, err = downloadUsingFileInfo(ctx, outputPath, forceOverwrite, repository, path, m)
	} else { // if we download the latest artifact (use path_is_aliased), we don't have all the data for the fileInfo struct, because no GET call was sent.
		tflog.Debug(ctx, "pathIsAliased == true")
		fileInfo, err = downloadWithoutChecks(ctx, outputPath, forceOverwrite, repository, path, m)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return packFileInfo(fileInfo, d)
}

func downloadUsingFileInfo(ctx context.Context, outputPath string, forceOverwrite bool, repository string, path string, m interface{}) (FileInfo, error) {
	fileInfo := FileInfo{}

	tflog.Debug(ctx, "Fetching file info", map[string]interface{}{
		"repository": repository,
		"path":       path,
	})

	client := m.(util.ProvderMetadata).Client
	// switch to using Sprintf because Resty's SetPathParams() escape the path
	// see https://github.com/go-resty/resty/blob/v2.7.0/middleware.go#L33
	// should use url.JoinPath() eventually in go 1.20
	requestUrl := fmt.Sprintf("%s/artifactory/api/storage/%s/%s", client.BaseURL, repository, path)
	_, err := client.R().
		SetResult(&fileInfo).
		Get(requestUrl)
	if err != nil {
		return fileInfo, err
	}

	tflog.Debug(ctx, "File info fetched", map[string]interface{}{
		"fileInfo": fileInfo,
	})

	checksumMatches := false
	fileExists := FileExists(outputPath)
	if fileExists {
		checksumMatches, err = VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to verify checksum for %s", outputPath))
			return fileInfo, err
		}
	}

	tflog.Debug(ctx, "File info checked", map[string]interface{}{
		"fileExists":      fileExists,
		"checksumMatches": checksumMatches,
	})

	/*--File Download logic--
	1. File doesn't exist
	2. In Data Source argument `force_overwrite` set to true, an existing file in the output_path will be overwritten. Ignore file exists or not
	3. File exists but check sum doesn't match
	*/
	if !fileExists || forceOverwrite || (fileExists && !checksumMatches) {
		tflog.Info(ctx, "Creating local output file")
		err := createOutputFile(outputPath)
		if err != nil {
			return fileInfo, err
		}
	} else { //download not required
		tflog.Info(ctx, "Skip downloading file")
		return fileInfo, nil
	}

	tflog.Debug(ctx, "Downloading file...", map[string]interface{}{
		"fileInfo.DownloadUri": fileInfo.DownloadUri,
		"outputPath":           outputPath,
	})
	_, err = m.(util.ProvderMetadata).Client.R().SetOutput(outputPath).Get(fileInfo.DownloadUri)
	if err != nil {
		return fileInfo, err
	}

	tflog.Debug(ctx, "Verify checksum with downloaded file")
	checksumMatches, err = VerifySha256Checksum(outputPath, fileInfo.Checksums.Sha256)
	if err != nil {
		return fileInfo, err
	}
	if !checksumMatches {
		return fileInfo, fmt.Errorf(
			"Checksums for file %s and %s do not match, expected %s",
			outputPath,
			fileInfo.DownloadUri,
			fileInfo.Checksums.Sha256,
		)
	}

	return fileInfo, nil
}

func downloadWithoutChecks(ctx context.Context, outputPath string, forceOverwrite bool, repository string, path string, m interface{}) (FileInfo, error) {
	fileInfo := FileInfo{
		Repo: repository,
		Path: path,
	}

	fileExists := FileExists(outputPath)

	tflog.Debug(ctx, "File info", map[string]interface{}{
		"fileInfo":   fileInfo,
		"fileExists": fileExists,
	})

	if !fileExists || forceOverwrite {
		tflog.Info(ctx, "Creating local output file")
		err := createOutputFile(outputPath)
		if err != nil {
			return fileInfo, err
		}
	} else { //download not required
		tflog.Info(ctx, "Skip downloading file")
		return fileInfo, nil
	}

	tflog.Debug(ctx, "Downloading file...", map[string]interface{}{
		"repository": repository,
		"path":       path,
		"outputPath": outputPath,
	})

	client := m.(util.ProvderMetadata).Client
	// switch to using Sprintf because Resty's SetPathParams() escape the path
	// see https://github.com/go-resty/resty/blob/v2.7.0/middleware.go#L33
	// should use url.JoinPath() eventually in go 1.20
	requestUrl := fmt.Sprintf("%s/artifactory/%s/%s", client.BaseURL, repository, path)
	_, err := client.R().
		SetOutput(outputPath).
		Get(requestUrl)
	if err != nil {
		return fileInfo, err
	}

	return fileInfo, nil
}
