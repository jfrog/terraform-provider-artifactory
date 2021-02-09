package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clientio "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	ARTIFACTORY_SYMLINK = "symlink.dest"
	SYMLINK_SHA1        = "symlink.destsha1"
	Latest              = "LATEST"
	LastRelease         = "LAST_RELEASE"
)

func UploadFile(localPath, url, logMsgPrefix string, artifactoryDetails *auth.ServiceDetails, details *fileutils.FileDetails,
	httpClientsDetails httputils.HttpClientDetails, client *rthttpclient.ArtifactoryHttpClient, retries int, progress clientio.Progress) (*http.Response, []byte, error) {
	var err error
	if details == nil {
		details, err = fileutils.GetFileDetails(localPath)
	}
	if err != nil {
		return nil, nil, err
	}

	requestClientDetails := httpClientsDetails.Clone()
	AddChecksumHeaders(requestClientDetails.Headers, details)
	AddAuthHeaders(requestClientDetails.Headers, *artifactoryDetails)

	return client.UploadFile(localPath, url, logMsgPrefix, requestClientDetails, retries, progress)
}

func AddChecksumHeaders(headers map[string]string, fileDetails *fileutils.FileDetails) {
	AddHeader("X-Checksum-Sha1", fileDetails.Checksum.Sha1, &headers)
	AddHeader("X-Checksum-Md5", fileDetails.Checksum.Md5, &headers)
	if len(fileDetails.Checksum.Sha256) > 0 {
		AddHeader("X-Checksum", fileDetails.Checksum.Sha256, &headers)
	}
}

func AddAuthHeaders(headers map[string]string, artifactoryDetails auth.ServiceDetails) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if artifactoryDetails.GetSshAuthHeaders() != nil {
		utils.MergeMaps(artifactoryDetails.GetSshAuthHeaders(), headers)
	}
}

func SetContentType(contentType string, headers *map[string]string) {
	AddHeader("Content-Type", contentType, headers)
}

func DisableAccelBuffering(headers *map[string]string) {
	AddHeader("X-Accel-Buffering", "no", headers)
}

func AddHeader(headerName, headerValue string, headers *map[string]string) {
	if *headers == nil {
		*headers = make(map[string]string)
	}
	(*headers)[headerName] = headerValue
}

func BuildArtifactoryUrl(baseUrl, path string, params map[string]string) (string, error) {
	u := url.URL{Path: path}
	escapedUrl, err := url.Parse(baseUrl + u.String())
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	q := escapedUrl.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	escapedUrl.RawQuery = q.Encode()
	return escapedUrl.String(), nil
}

func IsWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*") || strings.HasSuffix(pattern, "/") || !strings.Contains(pattern, "/")
}

// paths - Sorted array.
// index - Index of the current path which we want to check if it a prefix of any of the other previous paths.
// separator - File separator.
// Returns true paths[index] is a prefix of any of the paths[i] where i<index, otherwise returns false.
func IsSubPath(paths []string, index int, separator string) bool {
	currentPath := paths[index]
	if !strings.HasSuffix(currentPath, separator) {
		currentPath += separator
	}
	for i := index - 1; i >= 0; i-- {
		if strings.HasPrefix(paths[i], currentPath) {
			return true
		}
	}
	return false
}

// This method parses buildIdentifier. buildIdentifier should be from the format "buildName/buildNumber".
// If no buildNumber provided LATEST will be downloaded.
// If buildName or buildNumber contains "/" (slash) it should be escaped by "\" (backslash).
// Result examples of parsing: "aaa/123" > "aaa"-"123", "aaa" > "aaa"-"LATEST", "aaa\\/aaa" > "aaa/aaa"-"LATEST",  "aaa/12\\/3" > "aaa"-"12/3".
func getBuildNameAndNumberFromBuildIdentifier(buildIdentifier string, flags CommonConf) (string, string, error) {
	buildName, buildNumber, err := parseNameAndVersion(buildIdentifier, true)
	if err != nil {
		return "", "", err
	}
	return GetBuildNameAndNumberFromArtifactory(buildName, buildNumber, flags)
}

func GetBuildNameAndNumberFromArtifactory(buildName, buildNumber string, flags CommonConf) (string, string, error) {
	if buildNumber == Latest || buildNumber == LastRelease {
		return getLatestBuildNumberFromArtifactory(buildName, buildNumber, flags)
	}
	return buildName, buildNumber, nil
}

func getBuildNameAndNumberFromProps(properties []Property) (buildName string, buildNumber string) {
	for _, property := range properties {
		if property.Key == "build.name" {
			buildName = property.Value
		} else if property.Key == "build.number" {
			buildNumber = property.Value
		}
		if len(buildName) > 0 && len(buildNumber) > 0 {
			return buildName, buildNumber
		}
	}
	return
}

// For builds (useLatestPolicy = true) - Parse build name and number. The build number can be LATEST if absent.
// For release bundles - Parse bundle name and version.
func parseNameAndVersion(identifier string, useLatestPolicy bool) (string, string, error) {
	const Delimiter = "/"
	const EscapeChar = "\\"

	if identifier == "" {
		return "", "", nil
	}
	if !strings.Contains(identifier, Delimiter) {
		if useLatestPolicy {
			log.Debug("No '" + Delimiter + "' is found in the build, build number is set to " + Latest)
			return identifier, Latest, nil
		} else {
			return "", "", errorutils.CheckError(errors.New("No '" + Delimiter + "' is found in the bundle"))
		}
	}
	name, version := "", ""
	versionsArray := []string{}
	identifiers := strings.Split(identifier, Delimiter)
	// The delimiter must not be prefixed with escapeChar (if it is, it should be part of the version)
	// the code below gets substring from before the last delimiter.
	// If the new string ends with escape char it means the last delimiter was part of the version and we need
	// to go back to the previous delimiter.
	// If no proper delimiter was found the full string will be the name.
	for i := len(identifiers) - 1; i >= 1; i-- {
		versionsArray = append([]string{identifiers[i]}, versionsArray...)
		if !strings.HasSuffix(identifiers[i-1], EscapeChar) {
			name = strings.Join(identifiers[:i], Delimiter)
			version = strings.Join(versionsArray, Delimiter)
			break
		}
	}
	if name == "" {
		if useLatestPolicy {
			log.Debug("No delimiter char (" + Delimiter + ") without escaping char was found in the build, build number is set to " + Latest)
			name = identifier
			version = Latest
		} else {
			return "", "", errorutils.CheckError(errors.New("No delimiter char (" + Delimiter + ") without escaping char was found in the bundle"))
		}
	}
	// Remove escape chars.
	name = strings.Replace(name, "\\/", "/", -1)
	version = strings.Replace(version, "\\/", "/", -1)
	return name, version, nil
}

type build struct {
	BuildName   string `json:"buildName"`
	BuildNumber string `json:"buildNumber"`
}

func getLatestBuildNumberFromArtifactory(buildName, buildNumber string, flags CommonConf) (string, string, error) {
	restUrl := flags.GetArtifactoryDetails().GetUrl() + "api/build/patternArtifacts"
	body, err := createBodyForLatestBuildRequest(buildName, buildNumber)
	if err != nil {
		return "", "", err
	}
	log.Debug("Getting build name and number from Artifactory: " + buildName + ", " + buildNumber)
	httpClientsDetails := flags.GetArtifactoryDetails().CreateHttpClientDetails()
	SetContentType("application/json", &httpClientsDetails.Headers)
	log.Debug("Sending post request to: " + restUrl + ", with the following body: " + string(body))
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return "", "", err
	}
	resp, body, err := client.SendPost(restUrl, body, httpClientsDetails)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}
	log.Debug("Artifactory response: ", resp.Status)
	var responseBuild []build
	err = json.Unmarshal(body, &responseBuild)
	if errorutils.CheckError(err) != nil {
		return "", "", err
	}
	if responseBuild[0].BuildNumber != "" {
		log.Debug("Found build number: " + responseBuild[0].BuildNumber)
	} else {
		log.Debug("The build could not be found in Artifactory")
	}

	return buildName, responseBuild[0].BuildNumber, nil
}

func createBodyForLatestBuildRequest(buildName, buildNumber string) (body []byte, err error) {
	buildJsonArray := []build{{buildName, buildNumber}}
	body, err = json.Marshal(buildJsonArray)
	err = errorutils.CheckError(err)
	return
}

func filterAqlSearchResultsByBuild(specFile *ArtifactoryCommonParams, reader *content.ContentReader, flags CommonConf, itemsAlreadyContainProperties bool) (*content.ContentReader, error) {
	var aqlSearchErr error
	var readerWithProps *content.ContentReader
	var buildArtifactsSha1 map[string]int
	var wg sync.WaitGroup
	// If 'build-number' is missing in spec file, we fetch the laster from artifactory.
	buildName, buildNumber, err := getBuildNameAndNumberFromBuildIdentifier(specFile.Build, flags)
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	// Get Sha1 for artifacts by build name and number
	go func() {
		buildArtifactsSha1, aqlSearchErr = fetchBuildArtifactsSha1(buildName, buildNumber, flags)
		wg.Done()
	}()

	if !itemsAlreadyContainProperties {
		// Add properties to the previously found artifacts (in case properties haven't already fetched from Artifactory)
		readerWithProps, err = searchProps(specFile.Aql.ItemsFind, "build.name", buildName, flags)
		if err != nil {
			return nil, err
		}
		defer readerWithProps.Close()
		tempReader, err := loadMissingProperties(reader, readerWithProps)
		if err != nil {
			return nil, err
		}
		defer tempReader.Close()
		wg.Wait()
		if aqlSearchErr != nil {
			return nil, aqlSearchErr
		}
		return filterBuildAqlSearchResults(tempReader, buildArtifactsSha1, buildName, buildNumber)
	}

	wg.Wait()
	if aqlSearchErr != nil {
		return nil, aqlSearchErr
	}
	return filterBuildAqlSearchResults(reader, buildArtifactsSha1, buildName, buildNumber)
}

// Load all properties to the sorted result items. Save the new result items to a file.
// cr - Sorted result without properties
// crWithProps - Result item with properties
// Return a content reader which points to the result file.
func loadMissingProperties(reader *content.ContentReader, readerWithProps *content.ContentReader) (*content.ContentReader, error) {
	// Key -> Relative path, value -> ResultItem
	// Contains limited amount of items from a file, to not overflow memory.
	buffer := make(map[string]*ResultItem)
	var err error
	// Create new file to write result output
	resultFile, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer resultFile.Close()
	for resultItem := new(ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		// Save the item in a buffer.
		buffer[resultItem.GetItemRelativePath()] = resultItem
		if len(buffer) == utils.MaxBufferSize {
			// Buffer was full, write all data to a file.
			err = updateProps(readerWithProps, resultFile, buffer)
			if err != nil {
				return nil, err
			}
			// Init buffer.
			buffer = make(map[string]*ResultItem)
		}
	}
	if reader.GetError() != nil {
		return nil, err
	}
	reader.Reset()
	if err := updateProps(readerWithProps, resultFile, buffer); err != nil {
		return nil, err
	}
	return content.NewContentReader(resultFile.GetFilePath(), content.DefaultKey), nil
}

// Load the properties from readerWithProps into buffer's ResultItem, sort the buffers keys, and write its values into the resultWriter.
// buffer - Search result buffer (sorted) Key -> relative path, value -> ResultItem.
// crWithProps - File containing all the results with proprties.
// resultWriter - Sorted search result with props.
func updateProps(crWithProps *content.ContentReader, resultWriter *content.ContentWriter, buffer map[string]*ResultItem) error {
	if len(buffer) == 0 {
		return nil
	}
	// Load buffer items with their properties.
	for resultItem := new(ResultItem); crWithProps.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		if value, ok := buffer[resultItem.GetItemRelativePath()]; ok {
			value.Properties = resultItem.Properties
		}
	}
	if err := crWithProps.GetError(); err != nil {
		return err
	}
	crWithProps.Reset()
	// Write the items to a file sorted.
	keys := make([]string, 0)
	for k := range buffer {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		resultWriter.Write(*buffer[k])
	}
	return nil
}

// Run AQL to retrieve all artifacts associated with a specific build.
// Return a map of the artifacts SHA1.
func fetchBuildArtifactsSha1(buildName, buildNumber string, flags CommonConf) (map[string]int, error) {
	buildQuery := createAqlQueryForBuild(buildName, buildNumber, buildIncludeQueryPart([]string{"name", "repo", "path", "actual_sha1"}))
	reader, err := aqlSearch(buildQuery, flags)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return extractSha1FromAqlResponse(reader)
}

// Find artifacts with a specific property.
// aqlBody - AQL to execute together with property filter.
// filterByPropName - Property name to filter.
// filterByPropValue - Property value to filter.
// flags - Command flags for AQL execution.
func searchProps(aqlBody, filterByPropName, filterByPropValue string, flags CommonConf) (*content.ContentReader, error) {
	return ExecAqlSaveToFile(createPropsQuery(aqlBody, filterByPropName, filterByPropValue), flags)
}

// Gets a reader of AQL results, and return map with all the SHA1's as keys.
// The values for all the keys in the map is 2
func extractSha1FromAqlResponse(reader *content.ContentReader) (elementsMap map[string]int, err error) {
	elementsMap = make(map[string]int)
	for resultItem := new(ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		elementsMap[resultItem.Actual_Sha1] = 2
	}
	if err = reader.GetError(); err != nil {
		return
	}
	reader.Reset()
	return
}

// Returns a filtered search result file.
// Map each search result in one of three priority files:
// 1st priority: Match {Sha1, build name, build number}
// 2nd priority: Match {Sha1, build name}
// 3rd priority: Match {Sha1}
// As a result, any duplicated search result item will be split into a different priority list.
// Then merge all the priority list into a single file, so each item is present once in the result file according to the priority list.
// Side note: For each priority level, a single SHA1 can match multi artifacts under different modules.
// reader - Reader of the aql result.
// buildArtifactsSha - Map of all the build-name's sha1 as keys and int as its values. The int value represents priority wheres 0 is a high priority and 2 is lowest.
func filterBuildAqlSearchResults(reader *content.ContentReader, buildArtifactsSha map[string]int, buildName, buildNumber string) (*content.ContentReader, error) {
	priorityArray, err := createPrioritiesFiles()
	if err != nil {
		return nil, err
	}
	resultCw, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer resultCw.Close()
	// Step 1 - Fill the priority files with search results.
	for resultItem := new(ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
		if _, ok := buildArtifactsSha[resultItem.Actual_Sha1]; !ok {
			continue
		}
		resultBuildName, resultBuildNumber := getBuildNameAndNumberFromProps(resultItem.Properties)
		isBuildNameMatched := resultBuildName == buildName
		if isBuildNameMatched && resultBuildNumber == buildNumber {
			priorityArray[0].Write(*resultItem)
			buildArtifactsSha[resultItem.Actual_Sha1] = 0
			continue
		}
		if isBuildNameMatched && buildArtifactsSha[resultItem.Actual_Sha1] != 0 {
			priorityArray[1].Write(*resultItem)
			buildArtifactsSha[resultItem.Actual_Sha1] = 1
			continue
		}
		if buildArtifactsSha[resultItem.Actual_Sha1] == 2 {
			priorityArray[2].Write(*resultItem)
		}
	}
	if err = reader.GetError(); err != nil {
		return nil, err
	}
	reader.Reset()
	var priorityLevel int = 0
	// Step 2 - Append the files to the final results file.
	// Scan each priority artifacts and apply them to the final result, skip results that have been already written, by higher priority.
	for _, priority := range priorityArray {
		if err = priority.Close(); err != nil {
			return nil, err
		}
		temp := content.NewContentReader(priority.GetFilePath(), content.DefaultKey)
		for resultItem := new(ResultItem); temp.NextRecord(resultItem) == nil; resultItem = new(ResultItem) {
			if buildArtifactsSha[resultItem.Actual_Sha1] == priorityLevel {
				resultCw.Write(*resultItem)
			}
		}
		if err = temp.GetError(); err != nil {
			return nil, err
		}
		if err = temp.Close(); err != nil {
			return nil, err
		}
		priorityLevel++
	}
	return content.NewContentReader(resultCw.GetFilePath(), content.DefaultKey), nil
}

// Create priority files.
func createPrioritiesFiles() ([]*content.ContentWriter, error) {
	firstPriority, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	secondPriority, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	thirdPriority, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	return []*content.ContentWriter{firstPriority, secondPriority, thirdPriority}, nil
}

type CommonConf interface {
	GetArtifactoryDetails() auth.ServiceDetails
	SetArtifactoryDetails(rt auth.ServiceDetails)
	GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error)
	IsDryRun() bool
}

type CommonConfImpl struct {
	artDetails auth.ServiceDetails
	DryRun     bool
}

func (flags *CommonConfImpl) GetArtifactoryDetails() auth.ServiceDetails {
	return flags.artDetails
}

func (flags *CommonConfImpl) SetArtifactoryDetails(rt auth.ServiceDetails) {
	flags.artDetails = rt
}

func (flags *CommonConfImpl) IsDryRun() bool {
	return flags.DryRun
}

func (flags *CommonConfImpl) GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error) {
	return rthttpclient.ArtifactoryClientBuilder().SetServiceDetails(&flags.artDetails).Build()
}
