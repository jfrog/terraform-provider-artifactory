package services

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/jfrog/gofrog/parallel"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clientio "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mholt/archiver"
)

type DownloadService struct {
	client       *rthttpclient.ArtifactoryHttpClient
	Progress     clientio.Progress
	ArtDetails   auth.ServiceDetails
	DryRun       bool
	Threads      int
	ResultWriter *content.ContentWriter
}

func NewDownloadService(client *rthttpclient.ArtifactoryHttpClient) *DownloadService {
	return &DownloadService{client: client}
}

func (ds *DownloadService) GetArtifactoryDetails() auth.ServiceDetails {
	return ds.ArtDetails
}

func (ds *DownloadService) SetArtifactoryDetails(rt auth.ServiceDetails) {
	ds.ArtDetails = rt
}

func (ds *DownloadService) IsDryRun() bool {
	return ds.DryRun
}

func (ds *DownloadService) GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error) {
	return ds.client, nil
}

func (ds *DownloadService) GetThreads() int {
	return ds.Threads
}

func (ds *DownloadService) SetThreads(threads int) {
	ds.Threads = threads
}

func (ds *DownloadService) SetServiceDetails(artDetails auth.ServiceDetails) {
	ds.ArtDetails = artDetails
}

func (ds *DownloadService) SetDryRun(isDryRun bool) {
	ds.DryRun = isDryRun
}

func (ds *DownloadService) DownloadFiles(downloadParams ...DownloadParams) (int, int, error) {
	producerConsumer := parallel.NewBounedRunner(ds.GetThreads(), false)
	errorsQueue := clientutils.NewErrorsQueue(1)
	expectedChan := make(chan int, 1)
	successCounters := make([]int, ds.GetThreads())
	ds.prepareTasks(producerConsumer, expectedChan, successCounters, errorsQueue, downloadParams...)

	err := ds.performTasks(producerConsumer, errorsQueue)
	totalSuccess := 0
	for _, v := range successCounters {
		totalSuccess += v
	}
	return totalSuccess, <-expectedChan, err
}

func (ds *DownloadService) prepareTasks(producer parallel.Runner, expectedChan chan int, successCounters []int, errorsQueue *clientutils.ErrorsQueue, downloadParamsSlice ...DownloadParams) {
	go func() {
		defer producer.Done()
		defer close(expectedChan)
		totalTasks := 0
		// Iterate over file-spec groups and produce download tasks.
		// When encountering an error, log and move to next group.
		for _, downloadParams := range downloadParamsSlice {
			var err error
			var reader *content.ContentReader
			// Create handler function for the current group.
			fileHandlerFunc := ds.createFileHandlerFunc(downloadParams, successCounters)
			// Search items.
			log.Info("Searching items to download...")
			switch downloadParams.GetSpecType() {
			case utils.WILDCARD:
				reader, err = ds.collectFilesUsingWildcardPattern(downloadParams)
			case utils.BUILD:
				reader, err = utils.SearchBySpecWithBuild(downloadParams.GetFile(), ds)
			case utils.AQL:
				reader, err = utils.SearchBySpecWithAql(downloadParams.GetFile(), ds, utils.SYMLINK)
			}
			// Check for search errors.
			if err != nil {
				log.Error(err)
				errorsQueue.AddError(err)
				continue
			}
			// Produce download tasks for the download consumers.
			totalTasks += produceTasks(reader, downloadParams, producer, fileHandlerFunc, errorsQueue)
			reader.Close()
		}
		expectedChan <- totalTasks
	}()
}

func (ds *DownloadService) collectFilesUsingWildcardPattern(downloadParams DownloadParams) (*content.ContentReader, error) {
	return utils.SearchBySpecWithPattern(downloadParams.GetFile(), ds, utils.SYMLINK)
}

func produceTasks(reader *content.ContentReader, downloadParams DownloadParams, producer parallel.Runner, fileHandler fileHandlerFunc, errorsQueue *clientutils.ErrorsQueue) int {
	flat := downloadParams.IsFlat()
	// Collect all folders path which might be needed to create.
	// key = folder path, value = the necessary data for producing create folder task.
	directoriesData := make(map[string]DownloadData)
	// Store all the paths which was created implicitly due to file upload.
	alreadyCreatedDirs := make(map[string]bool)
	// Store all the keys of directoriesData as an array.
	var directoriesDataKeys []string
	// Task counter
	var tasksCount int
	for resultItem := new(utils.ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(utils.ResultItem) {
		tempData := DownloadData{
			Dependency:   *resultItem,
			DownloadPath: downloadParams.GetPattern(),
			Target:       downloadParams.GetTarget(),
			Flat:         flat,
		}
		if resultItem.Type != "folder" {
			// Add a task. A task is a function of type TaskFunc which later on will be executed by other go routine, the communication is done using channels.
			// The second argument is an error handling func in case the taskFunc return an error.
			tasksCount++
			producer.AddTaskWithError(fileHandler(tempData), errorsQueue.AddError)
			// We don't want to create directories which are created explicitly by download files when ArtifactoryCommonParams.IncludeDirs is used.
			alreadyCreatedDirs[resultItem.Path] = true
		} else {
			directoriesData, directoriesDataKeys = collectDirPathsToCreate(*resultItem, directoriesData, tempData, directoriesDataKeys)
		}
	}
	if err := reader.GetError(); err != nil {
		errorsQueue.AddError(errorutils.CheckError(err))
		return tasksCount
	}
	reader.Reset()
	addCreateDirsTasks(directoriesDataKeys, alreadyCreatedDirs, producer, fileHandler, directoriesData, errorsQueue, flat)
	return tasksCount
}

// Extract for the aqlResultItem the directory path, store the path the directoriesDataKeys and in the directoriesData map.
// In addition directoriesData holds the correlate DownloadData for each key, later on this DownloadData will be used to create a create dir tasks if needed.
// This function append the new data to directoriesDataKeys and to directoriesData and return the new map and the new []string
// We are storing all the keys of directoriesData in additional array(directoriesDataKeys) so we could sort the keys and access the maps in the sorted order.
func collectDirPathsToCreate(aqlResultItem utils.ResultItem, directoriesData map[string]DownloadData, tempData DownloadData, directoriesDataKeys []string) (map[string]DownloadData, []string) {
	key := aqlResultItem.Name
	if aqlResultItem.Path != "." {
		key = path.Join(aqlResultItem.Path, aqlResultItem.Name)
	}
	directoriesData[key] = tempData
	directoriesDataKeys = append(directoriesDataKeys, key)
	return directoriesData, directoriesDataKeys
}

func addCreateDirsTasks(directoriesDataKeys []string, alreadyCreatedDirs map[string]bool, producer parallel.Runner, fileHandler fileHandlerFunc, directoriesData map[string]DownloadData, errorsQueue *clientutils.ErrorsQueue, isFlat bool) {
	// Longest path first
	// We are going to create the longest path first by doing so all sub paths of the longest path will be created implicitly.
	sort.Sort(sort.Reverse(sort.StringSlice(directoriesDataKeys)))
	for index, v := range directoriesDataKeys {
		// In order to avoid duplication we need to check the path wasn't already created by the previous action.
		if v != "." && // For some files the returned path can be the root path, ".", in that case we doing need to create any directory.
			(index == 0 || !utils.IsSubPath(directoriesDataKeys, index, "/")) { // directoriesDataKeys store all the path which might needed to be created, that's include duplicated paths.
			// By sorting the directoriesDataKeys we can assure that the longest path was created and therefore no need to create all it's sub paths.

			// Some directories were created due to file download when we aren't in flat download flow.
			if isFlat {
				producer.AddTaskWithError(fileHandler(directoriesData[v]), errorsQueue.AddError)
			} else if !alreadyCreatedDirs[v] {
				producer.AddTaskWithError(fileHandler(directoriesData[v]), errorsQueue.AddError)
			}
		}
	}
	return
}

func (ds *DownloadService) performTasks(consumer parallel.Runner, errorsQueue *clientutils.ErrorsQueue) error {
	// Blocked until finish consuming
	consumer.Run()
	if ds.ResultWriter != nil {
		err := ds.ResultWriter.Close()
		if err != nil {
			return err
		}
	}
	return errorsQueue.GetError()
}

func createDependencyFileInfo(resultItem utils.ResultItem, localPath, localFileName string) utils.FileInfo {
	fileInfo := utils.FileInfo{
		ArtifactoryPath: resultItem.GetItemRelativePath(),
		FileHashes: &utils.FileHashes{
			Sha1: resultItem.Actual_Sha1,
			Md5:  resultItem.Actual_Md5,
		},
	}
	fileInfo.LocalPath = filepath.Join(localPath, localFileName)
	return fileInfo
}

func createDownloadFileDetails(downloadPath, localPath, localFileName string, downloadData DownloadData) (details *httpclient.DownloadFileDetails) {
	details = &httpclient.DownloadFileDetails{
		FileName:      downloadData.Dependency.Name,
		DownloadPath:  downloadPath,
		RelativePath:  downloadData.Dependency.GetItemRelativePath(),
		LocalPath:     localPath,
		LocalFileName: localFileName,
		Size:          downloadData.Dependency.Size,
		ExpectedSha1:  downloadData.Dependency.Actual_Sha1}
	return
}

func (ds *DownloadService) downloadFile(downloadFileDetails *httpclient.DownloadFileDetails, logMsgPrefix string, downloadParams DownloadParams) error {
	httpClientsDetails := ds.ArtDetails.CreateHttpClientDetails()
	bulkDownload := downloadParams.SplitCount == 0 || downloadParams.MinSplitSize < 0 || downloadParams.MinSplitSize*1000 > downloadFileDetails.Size
	if !bulkDownload {
		acceptRange, err := ds.isFileAcceptRange(downloadFileDetails)
		if err != nil {
			return err
		}
		bulkDownload = !acceptRange
	}
	if bulkDownload {
		var resp *http.Response
		resp, err := ds.client.DownloadFileWithProgress(downloadFileDetails, logMsgPrefix, &httpClientsDetails,
			downloadParams.GetRetries(), downloadParams.IsExplode(), ds.Progress)
		if err != nil {
			return err
		}
		log.Debug(logMsgPrefix, "Artifactory response:", resp.Status)
		return errorutils.CheckResponseStatus(resp, http.StatusOK)
	}

	concurrentDownloadFlags := httpclient.ConcurrentDownloadFlags{
		FileName:      downloadFileDetails.FileName,
		DownloadPath:  downloadFileDetails.DownloadPath,
		RelativePath:  downloadFileDetails.RelativePath,
		LocalFileName: downloadFileDetails.LocalFileName,
		LocalPath:     downloadFileDetails.LocalPath,
		ExpectedSha1:  downloadFileDetails.ExpectedSha1,
		FileSize:      downloadFileDetails.Size,
		SplitCount:    downloadParams.SplitCount,
		Explode:       downloadParams.IsExplode(),
		Retries:       downloadParams.GetRetries()}

	resp, err := ds.client.DownloadFileConcurrently(concurrentDownloadFlags, logMsgPrefix, &httpClientsDetails, ds.Progress)
	if err != nil {
		return err
	}
	return errorutils.CheckResponseStatus(resp, http.StatusPartialContent)
}

func (ds *DownloadService) isFileAcceptRange(downloadFileDetails *httpclient.DownloadFileDetails) (bool, error) {
	httpClientsDetails := ds.ArtDetails.CreateHttpClientDetails()
	isAcceptRange, resp, err := ds.client.IsAcceptRanges(downloadFileDetails.DownloadPath, &httpClientsDetails)
	if err != nil {
		return false, err
	}
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return false, err
	}
	return isAcceptRange, err
}

func shouldDownloadFile(localFilePath, md5, sha1 string) (bool, error) {
	exists, err := fileutils.IsFileExists(localFilePath, false)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	localFileDetails, err := fileutils.GetFileDetails(localFilePath)
	if err != nil {
		return false, err
	}
	return localFileDetails.Checksum.Md5 != md5 || localFileDetails.Checksum.Sha1 != sha1, nil
}

func removeIfSymlink(localSymlinkPath string) error {
	if fileutils.IsPathSymlink(localSymlinkPath) {
		if err := os.Remove(localSymlinkPath); errorutils.CheckError(err) != nil {
			return err
		}
	}
	return nil
}

func createLocalSymlink(localPath, localFileName, symlinkArtifact string, symlinkChecksum bool, symlinkContentChecksum string, logMsgPrefix string) error {
	if symlinkChecksum && symlinkContentChecksum != "" {
		if !fileutils.IsPathExists(symlinkArtifact, false) {
			return errorutils.CheckError(errors.New("Symlink validation failed, target doesn't exist: " + symlinkArtifact))
		}
		file, err := os.Open(symlinkArtifact)
		if err = errorutils.CheckError(err); err != nil {
			return err
		}
		defer file.Close()
		checksumInfo, err := checksum.Calc(file, checksum.SHA1)
		if err != nil {
			return err
		}
		sha1 := checksumInfo[checksum.SHA1]
		if sha1 != symlinkContentChecksum {
			return errorutils.CheckError(errors.New("Symlink validation failed for target: " + symlinkArtifact))
		}
	}
	localSymlinkPath := filepath.Join(localPath, localFileName)
	isFileExists, err := fileutils.IsFileExists(localSymlinkPath, false)
	if err != nil {
		return err
	}
	// We can't create symlink in case a file with the same name already exist, we must remove the file before creating the symlink
	if isFileExists {
		if err := os.Remove(localSymlinkPath); err != nil {
			return err
		}
	}
	// Need to prepare the directories hierarchy
	_, err = fileutils.CreateFilePath(localPath, localFileName)
	if err != nil {
		return err
	}
	err = os.Symlink(symlinkArtifact, localSymlinkPath)
	if errorutils.CheckError(err) != nil {
		return err
	}
	log.Debug(logMsgPrefix, "Creating symlink file.")
	return nil
}

func getArtifactPropertyByKey(properties []utils.Property, key string) string {
	for _, v := range properties {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

func getArtifactSymlinkPath(properties []utils.Property) string {
	return getArtifactPropertyByKey(properties, utils.ARTIFACTORY_SYMLINK)
}

func getArtifactSymlinkChecksum(properties []utils.Property) string {
	return getArtifactPropertyByKey(properties, utils.SYMLINK_SHA1)
}

type fileHandlerFunc func(DownloadData) parallel.TaskFunc

func (ds *DownloadService) createFileHandlerFunc(downloadParams DownloadParams, successCounters []int) fileHandlerFunc {
	return func(downloadData DownloadData) parallel.TaskFunc {
		return func(threadId int) error {
			logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, ds.DryRun)
			downloadPath, e := utils.BuildArtifactoryUrl(ds.ArtDetails.GetUrl(), downloadData.Dependency.GetItemRelativePath(), make(map[string]string))
			if e != nil {
				return e
			}
			log.Info(logMsgPrefix+"Downloading", downloadData.Dependency.GetItemRelativePath())
			if ds.DryRun {
				return nil
			}
			target, e := clientutils.BuildTargetPath(downloadData.DownloadPath, downloadData.Dependency.GetItemRelativePath(), downloadData.Target, true)
			if e != nil {
				return e
			}
			localPath, localFileName := fileutils.GetLocalPathAndFile(downloadData.Dependency.Name, downloadData.Dependency.Path, target, downloadData.Flat)
			if downloadData.Dependency.Type == "folder" {
				return createDir(localPath, localFileName, logMsgPrefix)
			}
			e = removeIfSymlink(filepath.Join(localPath, localFileName))
			if e != nil {
				return e
			}
			if downloadParams.IsSymlink() {
				if isSymlink, e := createSymlinkIfNeeded(localPath, localFileName, logMsgPrefix, downloadData, successCounters, ds.ResultWriter, threadId, downloadParams); isSymlink {
					return e
				}
			}
			dependency := createDependencyFileInfo(downloadData.Dependency, localPath, localFileName)
			e = ds.downloadFileIfNeeded(downloadPath, localPath, localFileName, logMsgPrefix, downloadData, downloadParams)
			if e != nil {
				log.Error(logMsgPrefix, "Received an error: "+e.Error())
				return e
			}
			successCounters[threadId]++
			if ds.ResultWriter != nil {
				ds.ResultWriter.Write(dependency)
			}
			return nil
		}
	}
}

func (ds *DownloadService) downloadFileIfNeeded(downloadPath, localPath, localFileName, logMsgPrefix string, downloadData DownloadData, downloadParams DownloadParams) error {
	shouldDownload, e := shouldDownloadFile(filepath.Join(localPath, localFileName), downloadData.Dependency.Actual_Md5, downloadData.Dependency.Actual_Sha1)
	if e != nil {
		return e
	}
	if !shouldDownload {
		log.Debug(logMsgPrefix, "File already exists locally.")
		if downloadParams.IsExplode() {
			e = explodeLocalFile(localPath, localFileName)
		}
		return e
	}
	downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, downloadData)
	return ds.downloadFile(downloadFileDetails, logMsgPrefix, downloadParams)
}

func explodeLocalFile(localPath, localFileName string) (err error) {
	log.Info("Extracting archive:", localFileName, "to", localPath)
	arch := archiver.MatchingFormat(localFileName)
	absolutePath := filepath.Join(localPath, localFileName)
	err = nil

	// The file is indeed an archive
	if arch != nil {
		err := arch.Open(absolutePath, localPath)
		if err != nil {
			return errorutils.CheckError(err)
		}
		// If the file was extracted successfully, remove it from the file system
		err = os.Remove(absolutePath)
	}

	return errorutils.CheckError(err)
}

func createDir(localPath, localFileName, logMsgPrefix string) error {
	folderPath := filepath.Join(localPath, localFileName)
	e := fileutils.CreateDirIfNotExist(folderPath)
	if e != nil {
		return e
	}
	log.Info(logMsgPrefix + "Creating folder: " + folderPath)
	return nil
}

func createSymlinkIfNeeded(localPath, localFileName, logMsgPrefix string, downloadData DownloadData, successCounters []int, responseWriter *content.ContentWriter, threadId int, downloadParams DownloadParams) (bool, error) {
	symlinkArtifact := getArtifactSymlinkPath(downloadData.Dependency.Properties)
	isSymlink := len(symlinkArtifact) > 0
	if isSymlink {
		symlinkChecksum := getArtifactSymlinkChecksum(downloadData.Dependency.Properties)
		if e := createLocalSymlink(localPath, localFileName, symlinkArtifact, downloadParams.ValidateSymlinks(), symlinkChecksum, logMsgPrefix); e != nil {
			return isSymlink, e
		}
		dependency := createDependencyFileInfo(downloadData.Dependency, localPath, localFileName)
		successCounters[threadId]++
		if responseWriter != nil {
			responseWriter.Write(dependency)
		}
		return isSymlink, nil
	}
	return isSymlink, nil
}

type DownloadData struct {
	Dependency   utils.ResultItem
	DownloadPath string
	Target       string
	Flat         bool
}

type DownloadParams struct {
	*utils.ArtifactoryCommonParams
	Symlink         bool
	ValidateSymlink bool
	Flat            bool
	Explode         bool
	MinSplitSize    int64
	SplitCount      int
	Retries         int
}

func (ds *DownloadParams) IsFlat() bool {
	return ds.Flat
}

func (ds *DownloadParams) IsExplode() bool {
	return ds.Explode
}

func (ds *DownloadParams) GetFile() *utils.ArtifactoryCommonParams {
	return ds.ArtifactoryCommonParams
}

func (ds *DownloadParams) IsSymlink() bool {
	return ds.Symlink
}

func (ds *DownloadParams) ValidateSymlinks() bool {
	return ds.ValidateSymlink
}

func (ds *DownloadParams) GetRetries() int {
	return ds.Retries
}

func NewDownloadParams() DownloadParams {
	return DownloadParams{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{}, MinSplitSize: 5120, SplitCount: 3, Retries: 3}
}
