package _go

import (
	"net/http"
	"net/url"
	"strings"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

func init() {
	register(&publishZipAndModApi{})
}

const ArtifactoryMinSupportedVersionForInfoFile = "6.10.0"

// Support for Artifactory 6.6.1 and above API
type publishZipAndModApi struct {
	artifactoryVersion string
	clientDetails      httputils.HttpClientDetails
	client             *rthttpclient.ArtifactoryHttpClient
}

func (pwa *publishZipAndModApi) isCompatible(artifactoryVersion string) bool {
	propertiesApi := "6.6.1"
	version := version.NewVersion(artifactoryVersion)
	pwa.artifactoryVersion = artifactoryVersion
	return version.AtLeast(propertiesApi)
}

func (pwa *publishZipAndModApi) PublishPackage(params GoParams, client *rthttpclient.ArtifactoryHttpClient, ArtDetails auth.ServiceDetails) error {
	url, err := utils.BuildArtifactoryUrl(ArtDetails.GetUrl(), "api/go/"+params.GetTargetRepo(), make(map[string]string))
	if err != nil {
		return err
	}
	pwa.clientDetails = ArtDetails.CreateHttpClientDetails()
	pwa.client = client
	moduleId := strings.Split(params.GetModuleId(), ":")
	// Upload zip file
	err = pwa.upload(params.GetZipPath(), moduleId[0], params.GetVersion(), params.GetProps(), ".zip", url)
	if err != nil {
		return err
	}
	// Upload mod file
	err = pwa.upload(params.GetModPath(), moduleId[0], params.GetVersion(), params.GetProps(), ".mod", url)
	if err != nil {
		return err
	}
	if version.NewVersion(pwa.artifactoryVersion).AtLeast(ArtifactoryMinSupportedVersionForInfoFile) && params.GetInfoPath() != "" {
		// Upload info file. This supported from Artifactory version 6.10.0 and above
		return pwa.upload(params.GetInfoPath(), moduleId[0], params.GetVersion(), params.GetProps(), ".info", url)
	}
	return nil
}

func addGoVersion(version string, urlPath *string) {
	*urlPath += ";go.version=" + url.QueryEscape(version)
}

// localPath - The location of the file on the file system.
// moduleId - The name of the module for example github.com/jfrog/jfrog-client-go.
// version - The version of the project that being uploaded.
// props - The properties to be assigned for each artifact
// ext - The extension of the file: zip, mod, info. This extension will be joined with the version for the path. For example v1.2.3.info or v1.2.3.zip
// urlPath - The url including the repository. For example: http://127.0.0.1/artifactory/api/go/go-local
func (pwa *publishZipAndModApi) upload(localPath, moduleId, version, props, ext, urlPath string) error {
	err := CreateUrlPath(moduleId, version, props, ext, &urlPath)
	if err != nil {
		return err
	}
	addGoVersion(version, &urlPath)
	details, err := fileutils.GetFileDetails(localPath)
	if err != nil {
		return err
	}
	utils.AddChecksumHeaders(pwa.clientDetails.Headers, details)
	resp, _, err := pwa.client.UploadFile(localPath, urlPath, "", &pwa.clientDetails, GoUploadRetries, nil)
	if err != nil {
		return err
	}
	return errorutils.CheckResponseStatus(resp, http.StatusCreated)
}
