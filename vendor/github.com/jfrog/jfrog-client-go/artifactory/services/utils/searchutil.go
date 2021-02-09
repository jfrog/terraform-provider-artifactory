package utils

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type RequiredArtifactProps int

// This enum defines which properties are required in the result of the aql.
// For example, when performing a copy/move command - the props are not needed, so we set RequiredArtifactProps to NONE.
const (
	ALL RequiredArtifactProps = iota
	SYMLINK
	NONE
)

// Use this function when searching by build without pattern or aql.
// Search with builds returns many results, some are not part of the build and others may be duplicated of the same artifact.
// 1. Save SHA1 values received for build-name.
// 2. Remove artifacts that not are present on the sha1 list
// 3. If we have more than one artifact with the same sha1:
// 	3.1 Compare the build-name & build-number among all the artifact with the same sha1.
// This will prevent unnecessary search upon all Artifactory:
func SearchBySpecWithBuild(specFile *ArtifactoryCommonParams, flags CommonConf) (*content.ContentReader, error) {
	buildName, buildNumber, err := getBuildNameAndNumberFromBuildIdentifier(specFile.Build, flags)
	if err != nil {
		return nil, err
	}
	specFile.Aql = Aql{ItemsFind: createAqlBodyForBuild(buildName, buildNumber)}
	executionQuery := BuildQueryFromSpecFile(specFile, ALL)
	reader, err := aqlSearch(executionQuery, flags)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// If artifacts' properties weren't fetched in previous aql, fetch now and add to results.
	if !includePropertiesInAqlForSpec(specFile) {
		readerWithProps, err := searchProps(specFile.Aql.ItemsFind, "build.name", buildName, flags)
		if err != nil {
			return nil, err
		}
		defer readerWithProps.Close()
		readerSortedWithProps, err := loadMissingProperties(reader, readerWithProps)
		if err != nil {
			return nil, err
		}
		buildArtifactsSha1, err := extractSha1FromAqlResponse(readerSortedWithProps)
		return filterBuildAqlSearchResults(readerSortedWithProps, buildArtifactsSha1, buildName, buildNumber)
	}

	buildArtifactsSha1, err := extractSha1FromAqlResponse(reader)
	return filterBuildAqlSearchResults(reader, buildArtifactsSha1, buildName, buildNumber)
}

// Perform search by pattern.
func SearchBySpecWithPattern(specFile *ArtifactoryCommonParams, flags CommonConf, requiredArtifactProps RequiredArtifactProps) (*content.ContentReader, error) {
	// Create AQL according to spec fields.
	query, err := CreateAqlBodyForSpecWithPattern(specFile)
	if err != nil {
		return nil, err
	}
	specFile.Aql = Aql{ItemsFind: query}
	return SearchBySpecWithAql(specFile, flags, requiredArtifactProps)
}

// Use this function when running Aql with pattern
func SearchBySpecWithAql(specFile *ArtifactoryCommonParams, flags CommonConf, requiredArtifactProps RequiredArtifactProps) (*content.ContentReader, error) {
	// Execute the search according to provided aql in specFile.
	var fetchedProps *content.ContentReader
	query := BuildQueryFromSpecFile(specFile, requiredArtifactProps)
	reader, err := aqlSearch(query, flags)
	if err != nil {
		return nil, err
	}
	filteredReader, err := FilterResultsByBuild(specFile, flags, requiredArtifactProps, reader)
	if err != nil {
		return nil, err
	}
	if filteredReader != nil {
		defer reader.Close()
		fetchedProps, err = fetchProps(specFile, flags, requiredArtifactProps, filteredReader)
		if fetchedProps != nil {
			defer filteredReader.Close()
			return fetchedProps, err
		}
		return filteredReader, err
	}
	fetchedProps, err = fetchProps(specFile, flags, requiredArtifactProps, reader)
	if fetchedProps != nil {
		defer reader.Close()
		return fetchedProps, err
	}
	return reader, err
}

// Filter the results by build, if no build found or items to filter, nil will be returned.
func FilterResultsByBuild(specFile *ArtifactoryCommonParams, flags CommonConf, requiredArtifactProps RequiredArtifactProps, reader *content.ContentReader) (*content.ContentReader, error) {
	length, err := reader.Length()
	if err != nil {
		return nil, err
	}
	if specFile.Build != "" && length > 0 {
		// If requiredArtifactProps is not NONE and 'includePropertiesInAqlForSpec' for specFile returned true, results contains properties for artifacts.
		resultsArtifactsIncludeProperties := requiredArtifactProps != NONE && includePropertiesInAqlForSpec(specFile)
		return filterAqlSearchResultsByBuild(specFile, reader, flags, resultsArtifactsIncludeProperties)
	}
	return nil, nil
}

// Fetch properties only if:
// 1. Properties weren't included in 'results'.
// AND
// 2. Properties weren't fetched during 'build' filtering
// Otherwise, nil will be returned
func fetchProps(specFile *ArtifactoryCommonParams, flags CommonConf, requiredArtifactProps RequiredArtifactProps, reader *content.ContentReader) (*content.ContentReader, error) {
	if !includePropertiesInAqlForSpec(specFile) && specFile.Build == "" && requiredArtifactProps != NONE {
		var readerWithProps *content.ContentReader
		var err error
		switch requiredArtifactProps {
		case ALL:
			readerWithProps, err = searchProps(specFile.Aql.ItemsFind, "*", "*", flags)
		case SYMLINK:
			readerWithProps, err = searchProps(specFile.Aql.ItemsFind, "symlink.dest", "*", flags)
		}
		if err != nil {
			return nil, err
		}
		defer readerWithProps.Close()
		return loadMissingProperties(reader, readerWithProps)
	}
	return nil, nil
}

func aqlSearch(aqlQuery string, flags CommonConf) (*content.ContentReader, error) {
	return ExecAqlSaveToFile(aqlQuery, flags)
}

func ExecAql(aqlQuery string, flags CommonConf) (io.ReadCloser, error) {
	client, err := flags.GetJfrogHttpClient()
	if err != nil {
		return nil, err
	}
	aqlUrl := flags.GetArtifactoryDetails().GetUrl() + "api/search/aql"
	log.Debug("Searching Artifactory using AQL query:\n", aqlQuery)
	httpClientsDetails := flags.GetArtifactoryDetails().CreateHttpClientDetails()
	resp, err := client.SendPostLeaveBodyOpen(aqlUrl, []byte(aqlQuery), &httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n"))
	}
	log.Debug("Artifactory response: ", resp.Status)
	return resp.Body, err
}

func ExecAqlSaveToFile(aqlQuery string, flags CommonConf) (*content.ContentReader, error) {
	body, err := ExecAql(aqlQuery, flags)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := body.Close()
		if err != nil {
			log.Warn("Could not close connection:" + err.Error() + ".")
		}
	}()
	log.Debug("Streaming data to file...")
	filePath, err := streamToFile(body)
	if err != nil {
		return nil, err
	}
	log.Debug("Finish streaming data successfully.")
	return content.NewContentReader(filePath, content.DefaultKey), err
}

// Save the reader output into a temp file.
// return the file path.
func streamToFile(reader io.Reader) (string, error) {
	var fd *os.File
	bufio := bufio.NewReaderSize(reader, 65536)
	fd, err := fileutils.CreateTempFile()
	if err != nil {
		return "", err
	}
	defer fd.Close()
	_, err = io.Copy(fd, bufio)
	return fd.Name(), errorutils.CheckError(err)
}

func LogSearchResults(numOfArtifacts int) {
	var msgSuffix = "artifacts."
	if numOfArtifacts == 1 {
		msgSuffix = "artifact."
	}
	log.Info("Found", strconv.Itoa(numOfArtifacts), msgSuffix)
}

type AqlSearchResult struct {
	Results []ResultItem
}

type ResultItem struct {
	Repo        string     `json:"repo,omitempty"`
	Path        string     `json:"path,omitempty"`
	Name        string     `json:"name,omitempty"`
	Actual_Md5  string     `json:"actual_md5,omitempty"`
	Actual_Sha1 string     `json:"actual_sha1,omitempty"`
	Size        int64      `json:"size,omitempty"`
	Created     string     `json:"created,omitempty"`
	Modified    string     `json:"modified,omitempty"`
	Properties  []Property `json:"properties,omitempty"`
	Type        string     `json:"type,omitempty"`
}

func (item ResultItem) GetItemRelativePath() string {
	if item.Path == "." {
		return path.Join(item.Repo, item.Name)
	}

	url := item.Repo
	url = addSeparator(url, "/", item.Path)
	url = addSeparator(url, "/", item.Name)
	if item.Type == "folder" && !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url
}

func addSeparator(str1, separator, str2 string) string {
	if str2 == "" {
		return str1
	}
	if str1 == "" {
		return str2
	}

	return str1 + separator + str2
}

func (item *ResultItem) ToArtifact() buildinfo.Artifact {
	return buildinfo.Artifact{Name: item.Name, Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}, Path: path.Join(item.Repo, item.Path, item.Name)}
}

func (item *ResultItem) ToDependency() buildinfo.Dependency {
	return buildinfo.Dependency{Id: item.Name, Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
}

type AqlSearchResultItemFilter func(*content.ContentReader) (*content.ContentReader, error)

func FilterBottomChainResults(reader *content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	var temp string
	for resultItem := new(ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		rPath := resultItem.GetItemRelativePath()
		if resultItem.Type == "folder" && !strings.HasSuffix(rPath, "/") {
			rPath += "/"
		}
		if temp == "" || !strings.HasPrefix(temp, rPath) {
			writer.Write(*resultItem)
			temp = rPath
		}
	}
	if err := reader.GetError(); err != nil {
		return nil, err
	}
	reader.Reset()
	return content.NewContentReader(writer.GetFilePath(), writer.GetArrayKey()), nil
}

// Reduce the amount of items by saveing only the shortest item path for each unique path e.g.:
// a | a/b | c | e/f -> a | c | e/f
func FilterTopChainResults(reader *content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	var prevFolder string
	for resultItem := new(ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		rPath := resultItem.GetItemRelativePath()
		if resultItem.Type == "folder" && !strings.HasSuffix(rPath, "/") {
			rPath += "/"
		}
		if prevFolder == "" || !strings.HasPrefix(rPath, prevFolder) {
			writer.Write(*resultItem)
			if resultItem.Type == "folder" {
				prevFolder = rPath
			}
		}
	}
	if err := reader.GetError(); err != nil {
		return nil, err
	}
	reader.Reset()
	return content.NewContentReader(writer.GetFilePath(), writer.GetArrayKey()), nil
}

func ReduceTopChainDirResult(searchResults *content.ContentReader) (*content.ContentReader, error) {
	return ReduceDirResult(searchResults, true, FilterTopChainResults)
}

func ReduceBottomChainDirResult(searchResults *content.ContentReader) (*content.ContentReader, error) {
	return ReduceDirResult(searchResults, false, FilterBottomChainResults)
}

// Reduce Dir results by using the resultsFilter
func ReduceDirResult(searchResults *content.ContentReader, ascendingOrder bool, resultsFilter AqlSearchResultItemFilter) (*content.ContentReader, error) {
	// Sort results in asc order according to relative path.
	// Split to files if the total result is bigget than the maximum buffest.
	paths := make(map[string]ResultItem)
	pathsKeys := make([]string, 0, utils.MaxBufferSize)
	sortedFiles := []*content.ContentReader{}
	defer func() {
		for _, file := range sortedFiles {
			file.Close()
		}
	}()
	for resultItem := new(ResultItem); searchResults.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		if resultItem.Name == "." {
			continue
		}
		rPath := resultItem.GetItemRelativePath()
		paths[rPath] = *resultItem
		pathsKeys = append(pathsKeys, rPath)
		if len(pathsKeys) == utils.MaxBufferSize {
			sortedFile, err := SortAndSaveBufferToFile(paths, pathsKeys, ascendingOrder)
			if err != nil {
				return nil, err
			}
			sortedFiles = append(sortedFiles, sortedFile)
			paths = make(map[string]ResultItem)
			pathsKeys = make([]string, 0, utils.MaxBufferSize)
		}
	}
	if err := searchResults.GetError(); err != nil {
		return nil, err
	}
	searchResults.Reset()
	var sortedFile *content.ContentReader
	if len(pathsKeys) > 0 {
		sortedFile, err := SortAndSaveBufferToFile(paths, pathsKeys, ascendingOrder)
		if err != nil {
			return nil, err
		}
		sortedFiles = append(sortedFiles, sortedFile)
	}
	// Merge sorted files
	sortedFile, err := MergeSortedFiles(sortedFiles, ascendingOrder)
	if err != nil {
		return nil, err
	}
	defer sortedFile.Close()
	return resultsFilter(sortedFile)
}

func SortAndSaveBufferToFile(paths map[string]ResultItem, pathsKeys []string, increasingOrder bool) (*content.ContentReader, error) {
	if len(pathsKeys) == 0 {
		return nil, nil
	}
	writer, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	if increasingOrder {
		sort.Strings(pathsKeys)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(pathsKeys)))
	}
	for _, v := range pathsKeys {
		writer.Write(paths[v])
	}
	return content.NewContentReader(writer.GetFilePath(), writer.GetArrayKey()), nil
}

// Merge all the sorted files into a single sorted file.
func MergeSortedFiles(sortedFiles []*content.ContentReader, ascendingOrder bool) (*content.ContentReader, error) {
	if len(sortedFiles) == 0 {
		return content.NewEmptyContentReader(content.DefaultKey), nil
	}
	resultWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer resultWriter.Close()
	currentResultItem := make([]*ResultItem, len(sortedFiles))
	sortedFilesClone := make([]*content.ContentReader, len(sortedFiles))
	copy(sortedFilesClone, sortedFiles)
	for {
		var candidateToWrite *ResultItem
		smallestIndex := 0
		for i := 0; i < len(sortedFilesClone); i++ {
			if currentResultItem[i] == nil && sortedFilesClone[i] != nil {
				temp := new(ResultItem)
				if err := sortedFilesClone[i].NextRecord(temp); nil != err {
					sortedFilesClone[i] = nil
					continue
				}
				currentResultItem[i] = temp
			}
			if candidateToWrite == nil || (currentResultItem[i] != nil && compareStrings(candidateToWrite.GetItemRelativePath(), currentResultItem[i].GetItemRelativePath(), ascendingOrder)) {
				candidateToWrite = currentResultItem[i]
				smallestIndex = i
			}
		}
		if candidateToWrite == nil {
			break
		}
		resultWriter.Write(*candidateToWrite)
		currentResultItem[smallestIndex] = nil
	}
	return content.NewContentReader(resultWriter.GetFilePath(), resultWriter.GetArrayKey()), nil
}

func compareStrings(src, against string, ascendingOrder bool) bool {
	if ascendingOrder {
		return src > against
	}
	return src < against
}
