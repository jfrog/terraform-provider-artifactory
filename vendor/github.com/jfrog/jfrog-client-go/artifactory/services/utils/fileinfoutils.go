package utils

import (
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"strings"
)

type FileHashes struct {
	Sha256 string `json:"sha256,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Md5    string `json:"md5,omitempty"`
}

type FileInfo struct {
	*FileHashes
	LocalPath               string `json:"localPath,omitempty"`
	ArtifactoryPath         string `json:"artifactoryPath,omitempty"`
	InternalArtifactoryPath string `json:"internalArtifactoryPath,omitempty"`
}

func (fileInfo *FileInfo) ToBuildArtifacts() buildinfo.Artifact {
	artifact := buildinfo.Artifact{Checksum: &buildinfo.Checksum{}}
	artifact.Sha1 = fileInfo.Sha1
	artifact.Md5 = fileInfo.Md5
	// Artifact name in build info as the name in artifactory
	filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.ArtifactoryPath)
	artifact.Name = filename
	if i := strings.LastIndex(filename, "."); i != -1 {
		artifact.Type = filename[i+1:]
	}
	artifact.Path = fileInfo.InternalArtifactoryPath
	return artifact
}

func FlattenFileInfoArray(dependenciesBuildInfo [][]FileInfo) []FileInfo {
	var buildInfo []FileInfo
	for _, v := range dependenciesBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}
