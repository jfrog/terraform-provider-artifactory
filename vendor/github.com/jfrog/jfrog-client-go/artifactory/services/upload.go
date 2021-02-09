package services

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jfrog/gofrog/parallel"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/fspatterns"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	ioutils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type UploadService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	Progress   ioutils.Progress
	ArtDetails auth.ServiceDetails
	DryRun     bool
	Threads    int
}

func NewUploadService(client *rthttpclient.ArtifactoryHttpClient) *UploadService {
	return &UploadService{client: client}
}

func (us *UploadService) SetThreads(threads int) {
	us.Threads = threads
}

func (us *UploadService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return us.client
}

func (us *UploadService) SetServiceDetails(artDetails auth.ServiceDetails) {
	us.ArtDetails = artDetails
}

func (us *UploadService) SetDryRun(isDryRun bool) {
	us.DryRun = isDryRun
}

func (us *UploadService) UploadFiles(uploadParams ...UploadParams) (artifactsFileInfo []utils.FileInfo, totalUploaded, totalFailed int, err error) {
	// Uploading threads are using this struct to report upload results.
	uploadSummary := *utils.NewUploadResult(us.Threads)
	producerConsumer := parallel.NewRunner(us.Threads, 100, false)
	errorsQueue := clientutils.NewErrorsQueue(1)
	us.prepareUploadTasks(producerConsumer, errorsQueue, uploadSummary, uploadParams...)
	return us.performUploadTasks(producerConsumer, &uploadSummary, errorsQueue)
}

func (us *UploadService) prepareUploadTasks(producer parallel.Runner, errorsQueue *clientutils.ErrorsQueue, uploadSummary utils.UploadResult, uploadParamsSlice ...UploadParams) {
	go func() {
		defer producer.Done()
		// Iterate over file-spec groups and produce upload tasks.
		// When encountering an error, log and move to next group.
		vcsCache := clientutils.NewVcsDetals()
		for _, uploadParams := range uploadParamsSlice {
			artifactHandlerFunc := us.createArtifactHandlerFunc(&uploadSummary, uploadParams)
			err := collectFilesForUpload(uploadParams, producer, artifactHandlerFunc, errorsQueue, vcsCache)
			if err != nil {
				log.Error(err)
				errorsQueue.AddError(err)
			}
		}
	}()
}

func (us *UploadService) performUploadTasks(consumer parallel.Runner, uploadSummary *utils.UploadResult, errorsQueue *clientutils.ErrorsQueue) (artifactsFileInfo []utils.FileInfo, totalUploaded, totalFailed int, err error) {
	// Blocking until consuming is finished.
	consumer.Run()
	err = errorsQueue.GetError()

	totalUploaded = utils.SumIntArray(uploadSummary.SuccessCount)
	totalUploadAttempted := utils.SumIntArray(uploadSummary.TotalCount)

	log.Debug("Uploaded", strconv.Itoa(totalUploaded), "artifacts.")
	totalFailed = totalUploadAttempted - totalUploaded
	if totalFailed > 0 {
		log.Error("Failed uploading", strconv.Itoa(totalFailed), "artifacts.")
	}
	artifactsFileInfo = utils.FlattenFileInfoArray(uploadSummary.FileInfo)
	return
}

func addProps(oldProps, additionalProps string) string {
	if len(oldProps) > 0 && !strings.HasSuffix(oldProps, ";") && len(additionalProps) > 0 {
		oldProps += ";"
	}
	return oldProps + additionalProps
}

func addSymlinkProps(artifact clientutils.Artifact, uploadParams UploadParams) (string, error) {
	artifactProps := ""
	artifactSymlink := artifact.Symlink
	if uploadParams.IsSymlink() && len(artifactSymlink) > 0 {
		sha1Property := ""
		fileInfo, err := os.Stat(artifact.LocalPath)
		if err != nil {
			// If error occurred, but not due to nonexistence of Symlink target -> return empty
			if !os.IsNotExist(err) {
				return "", err
			}
			// If Symlink target exists -> get SHA1 if isn't a directory
		} else if !fileInfo.IsDir() {
			file, err := os.Open(artifact.LocalPath)
			if err != nil {
				return "", errorutils.CheckError(err)
			}
			defer file.Close()
			checksumInfo, err := checksum.Calc(file, checksum.SHA1)
			if err != nil {
				return "", err
			}
			sha1 := checksumInfo[checksum.SHA1]
			sha1Property = ";" + utils.SYMLINK_SHA1 + "=" + sha1
		}
		artifactProps += utils.ARTIFACTORY_SYMLINK + "=" + artifactSymlink + sha1Property
	}
	props := uploadParams.GetProps()
	artifactProps = addProps(props, artifactProps)
	return artifactProps, nil
}

func collectFilesForUpload(uploadParams UploadParams, producer parallel.Runner, artifactHandlerFunc artifactContext, errorsQueue *clientutils.ErrorsQueue, vcsCache *clientutils.VcsCache) error {
	if strings.Index(uploadParams.GetTarget(), "/") < 0 {
		uploadParams.SetTarget(uploadParams.GetTarget() + "/")
	}
	uploadParams.SetPattern(clientutils.ReplaceTildeWithUserHome(uploadParams.GetPattern()))
	// Save parentheses index in pattern, witch have corresponding placeholder.
	rootPath, err := fspatterns.GetRootPath(uploadParams.GetPattern(), uploadParams.GetTarget(), uploadParams.IsRegexp(), uploadParams.IsSymlink())
	if err != nil {
		return err
	}

	isDir, err := fileutils.IsDirExists(rootPath, uploadParams.IsSymlink())
	if err != nil {
		return err
	}

	// If the path is a single file (or a symlink while preserving symlinks) upload it and return
	if !isDir || (fileutils.IsPathSymlink(rootPath) && uploadParams.IsSymlink()) {
		artifact, err := fspatterns.GetSingleFileToUpload(rootPath, uploadParams.GetTarget(), uploadParams.IsFlat(), uploadParams.IsSymlink())
		if err != nil {
			return err
		}
		props, err := addSymlinkProps(artifact, uploadParams)
		if err != nil {
			return err
		}
		if uploadParams.IsAddVcsProps() {
			vcsProps, err := getVcsProps(artifact.LocalPath, vcsCache)
			if err != nil {
				return err
			}
			props += vcsProps
		}
		uploadData := UploadData{Artifact: artifact, Props: props}
		task := artifactHandlerFunc(uploadData)
		producer.AddTaskWithError(task, errorsQueue.AddError)
		return err
	}
	uploadParams.SetPattern(clientutils.PrepareLocalPathForUpload(uploadParams.GetPattern(), uploadParams.IsRegexp()))
	err = collectPatternMatchingFiles(uploadParams, rootPath, producer, artifactHandlerFunc, errorsQueue, vcsCache)
	return err
}

func collectPatternMatchingFiles(uploadParams UploadParams, rootPath string, producer parallel.Runner, artifactHandlerFunc artifactContext, errorsQueue *clientutils.ErrorsQueue, vcsCache *clientutils.VcsCache) error {
	excludePathPattern := fspatterns.PrepareExcludePathPattern(uploadParams)
	patternRegex, err := regexp.Compile(uploadParams.GetPattern())
	if errorutils.CheckError(err) != nil {
		return err
	}

	paths, err := fspatterns.GetPaths(rootPath, uploadParams.IsRecursive(), uploadParams.IsIncludeDirs(), uploadParams.IsSymlink())
	if err != nil {
		return err
	}
	// Longest paths first
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	// 'foldersPaths' is a subset of the 'paths' array. foldersPaths is in use only when we need to upload folders with flat=true.
	// 'foldersPaths' will contain only the directories paths which are in the 'paths' array.
	var foldersPaths []string
	for index, path := range paths {
		matches, isDir, isSymlinkFlow, err := fspatterns.PrepareAndFilterPaths(path, excludePathPattern, uploadParams.IsSymlink(), uploadParams.IsIncludeDirs(), patternRegex)
		if err != nil {
			return err
		}

		if matches != nil && len(matches) > 0 {
			target := uploadParams.GetTarget()
			tempPaths := paths
			tempIndex := index
			// In case we need to upload directories with flat=true, we want to avoid the creation of unnecessary paths in Artifactory.
			// To achieve this, we need to take into consideration the directories which had already been uploaded, ignoring all files paths.
			// When flat=false we take into consideration folder paths which were created implicitly by file upload
			if uploadParams.IsFlat() && uploadParams.IsIncludeDirs() && isDir {
				foldersPaths = append(foldersPaths, path)
				tempPaths = foldersPaths
				tempIndex = len(foldersPaths) - 1
			}
			taskData := &uploadTaskData{target: target, path: path, isDir: isDir, isSymlinkFlow: isSymlinkFlow,
				paths: tempPaths, groups: matches, index: tempIndex, size: len(matches), uploadParams: uploadParams,
				producer: producer, artifactHandlerFunc: artifactHandlerFunc, errorsQueue: errorsQueue,
			}
			createUploadTask(taskData, vcsCache)
		}
	}
	return nil
}

type uploadTaskData struct {
	target              string
	path                string
	isDir               bool
	isSymlinkFlow       bool
	paths               []string
	groups              []string
	index               int
	size                int
	uploadParams        UploadParams
	producer            parallel.Runner
	artifactHandlerFunc artifactContext
	errorsQueue         *clientutils.ErrorsQueue
}

func createUploadTask(taskData *uploadTaskData, vcsCache *clientutils.VcsCache) error {
	for i := 1; i < taskData.size; i++ {
		group := strings.Replace(taskData.groups[i], "\\", "/", -1)
		taskData.target = strings.Replace(taskData.target, "{"+strconv.Itoa(i)+"}", group, -1)
	}
	var task parallel.TaskFunc

	// Get symlink target (returns empty string if regular file) - Used in upload name / symlinks properties
	symlinkPath, err := fspatterns.GetFileSymlinkPath(taskData.path)
	if err != nil {
		return err
	}

	// If preserving symlinks or symlink target is empty, use root path name for upload (symlink itself / regular file)
	if taskData.uploadParams.IsSymlink() || symlinkPath == "" {
		taskData.target = getUploadTarget(taskData.path, taskData.target, taskData.uploadParams.IsFlat())
	} else {
		taskData.target = getUploadTarget(symlinkPath, taskData.target, taskData.uploadParams.IsFlat())
	}

	artifact := clientutils.Artifact{LocalPath: taskData.path, TargetPath: taskData.target, Symlink: symlinkPath}
	props, e := addSymlinkProps(artifact, taskData.uploadParams)
	if e != nil {
		return e
	}
	if taskData.uploadParams.IsAddVcsProps() {
		vcsProps, err := getVcsProps(taskData.path, vcsCache)
		if err != nil {
			return err
		}
		props += vcsProps
	}
	uploadData := UploadData{Artifact: artifact, Props: props}
	if taskData.isDir && taskData.uploadParams.IsIncludeDirs() && !taskData.isSymlinkFlow {
		if taskData.path != "." && (taskData.index == 0 || !utils.IsSubPath(taskData.paths, taskData.index, fileutils.GetFileSeparator())) {
			uploadData.IsDir = true
		} else {
			return nil
		}
	}
	task = taskData.artifactHandlerFunc(uploadData)
	taskData.producer.AddTaskWithError(task, taskData.errorsQueue.AddError)
	return nil
}

// Construct the target path while taking `flat` flag into account.
func getUploadTarget(rootPath, target string, isFlat bool) string {
	if strings.HasSuffix(target, "/") {
		if isFlat {
			fileName, _ := fileutils.GetFileAndDirFromPath(rootPath)
			target += fileName
		} else {
			target += clientutils.TrimPath(rootPath)
		}
	}
	return target
}

func addPropsToTargetPath(targetPath, props, debConfig string) (string, error) {
	propsStr := strings.Join([]string{props, getDebianProps(debConfig)}, ";")
	properties, err := utils.ParseProperties(propsStr, utils.SplitCommas)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{targetPath, properties.ToEncodedString()}, ";"), nil
}

func prepareUploadData(localPath, baseTargetPath, props string, uploadParams UploadParams, logMsgPrefix string) (fileInfo os.FileInfo, targetPath string, err error) {
	targetPath, err = addPropsToTargetPath(baseTargetPath, props, uploadParams.GetDebian())
	if errorutils.CheckError(err) != nil {
		return
	}
	log.Info(logMsgPrefix+"Uploading artifact:", localPath)

	fileInfo, err = os.Lstat(localPath)
	errorutils.CheckError(err)
	return
}

// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func (us *UploadService) uploadFile(localPath, targetPath, pathInArtifactory, props string, uploadParams UploadParams, logMsgPrefix string) (utils.FileInfo, bool, error) {
	fileInfo, targetPathWithProps, err := prepareUploadData(localPath, targetPath, props, uploadParams, logMsgPrefix)
	if err != nil {
		return utils.FileInfo{}, false, err
	}

	var checksumDeployed = false
	var resp *http.Response
	var details *fileutils.FileDetails
	var body []byte
	httpClientsDetails := us.ArtDetails.CreateHttpClientDetails()
	if errorutils.CheckError(err) != nil {
		return utils.FileInfo{}, false, err
	}
	if uploadParams.IsSymlink() && fileutils.IsFileSymlink(fileInfo) {
		resp, details, body, err = us.uploadSymlink(targetPathWithProps, logMsgPrefix, httpClientsDetails, uploadParams)
	} else {
		resp, details, body, checksumDeployed, err = us.doUpload(localPath, targetPathWithProps, logMsgPrefix, httpClientsDetails, fileInfo, uploadParams)
	}
	if err != nil {
		return utils.FileInfo{}, false, err
	}
	logUploadResponse(logMsgPrefix, resp, body, checksumDeployed, us.DryRun)
	artifact := createBuildArtifactItem(details, localPath, targetPath, pathInArtifactory)
	return artifact, us.DryRun || checksumDeployed || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK, nil
}

func (us *UploadService) uploadSymlink(targetPath, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails, uploadParams UploadParams) (resp *http.Response, details *fileutils.FileDetails, body []byte, err error) {
	details, err = fspatterns.CreateSymlinkFileDetails()
	if err != nil {
		return
	}
	resp, body, err = utils.UploadFile("", targetPath, logMsgPrefix, &us.ArtDetails, details, httpClientsDetails, us.client, uploadParams.GetRetries(), nil)
	return
}

func (us *UploadService) doUpload(localPath, targetPath, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails, fileInfo os.FileInfo, uploadParams UploadParams) (*http.Response, *fileutils.FileDetails, []byte, bool, error) {
	var details *fileutils.FileDetails
	var checksumDeployed bool
	var resp *http.Response
	var body []byte
	var err error
	addExplodeHeader(&httpClientsDetails, uploadParams.IsExplodeArchive())
	if fileInfo.Size() >= uploadParams.MinChecksumDeploy && !uploadParams.IsExplodeArchive() {
		resp, details, body, err = us.tryChecksumDeploy(localPath, targetPath, httpClientsDetails, us.client)
		if err != nil {
			return resp, details, body, checksumDeployed, err
		}
		checksumDeployed = !us.DryRun && (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK)
	}
	if !us.DryRun && !checksumDeployed {
		var body []byte
		resp, body, err = utils.UploadFile(localPath, targetPath, logMsgPrefix, &us.ArtDetails, details,
			httpClientsDetails, us.client, uploadParams.Retries, us.Progress)
		if err != nil {
			return resp, details, body, checksumDeployed, err
		}
	}
	if details == nil {
		details, err = fileutils.GetFileDetails(localPath)
	}
	return resp, details, body, checksumDeployed, err
}

func logUploadResponse(logMsgPrefix string, resp *http.Response, body []byte, checksumDeployed, isDryRun bool) {
	if resp != nil && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Error(logMsgPrefix + "Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
		return
	}
	if !isDryRun {
		var strChecksumDeployed string
		if checksumDeployed {
			strChecksumDeployed = " (Checksum deploy)"
		} else {
			strChecksumDeployed = ""
		}
		log.Debug(logMsgPrefix, "Artifactory response:", resp.Status, strChecksumDeployed)
	}
}

func createBuildArtifactItem(details *fileutils.FileDetails, localPath, targetPath, pathInArtifactory string) utils.FileInfo {
	return utils.FileInfo{
		LocalPath:               localPath,
		ArtifactoryPath:         targetPath,
		InternalArtifactoryPath: pathInArtifactory,
		FileHashes: &utils.FileHashes{
			Sha256: details.Checksum.Sha256,
			Sha1:   details.Checksum.Sha1,
			Md5:    details.Checksum.Md5,
		},
	}
}

func addExplodeHeader(httpClientsDetails *httputils.HttpClientDetails, isExplode bool) {
	if isExplode {
		utils.AddHeader("X-Explode-Archive", "true", &httpClientsDetails.Headers)
	}
}

func (us *UploadService) tryChecksumDeploy(filePath, targetPath string, httpClientsDetails httputils.HttpClientDetails,
	client *rthttpclient.ArtifactoryHttpClient) (resp *http.Response, details *fileutils.FileDetails, body []byte, err error) {
	if us.DryRun {
		return
	}
	details, err = fileutils.GetFileDetails(filePath)
	if err != nil {
		return
	}

	requestClientDetails := httpClientsDetails.Clone()
	utils.AddHeader("X-Checksum-Deploy", "true", &requestClientDetails.Headers)
	utils.AddChecksumHeaders(requestClientDetails.Headers, details)
	utils.AddAuthHeaders(requestClientDetails.Headers, us.ArtDetails)

	resp, body, err = client.SendPut(targetPath, nil, requestClientDetails)
	return
}

func getDebianProps(debianPropsStr string) string {
	if debianPropsStr == "" {
		return ""
	}
	result := ""
	debProps := clientutils.SplitWithEscape(debianPropsStr, '/')
	for k, v := range []string{"deb.distribution", "deb.component", "deb.architecture"} {
		debProp := strings.Join([]string{v, debProps[k]}, "=")
		result = strings.Join([]string{result, debProp}, ";")
	}
	return result
}

type UploadParams struct {
	*utils.ArtifactoryCommonParams
	Deb               string
	Symlink           bool
	ExplodeArchive    bool
	Flat              bool
	AddVcsProps       bool
	Retries           int
	MinChecksumDeploy int64
}

func (up *UploadParams) IsFlat() bool {
	return up.Flat
}

func (up *UploadParams) IsSymlink() bool {
	return up.Symlink
}

func (up *UploadParams) IsAddVcsProps() bool {
	return up.AddVcsProps
}

func (up *UploadParams) IsExplodeArchive() bool {
	return up.ExplodeArchive
}

func (up *UploadParams) GetDebian() string {
	return up.Deb
}

func (up *UploadParams) GetRetries() int {
	return up.Retries
}

type UploadData struct {
	Artifact clientutils.Artifact
	Props    string
	IsDir    bool
}

type artifactContext func(UploadData) parallel.TaskFunc

func (us *UploadService) createArtifactHandlerFunc(uploadResult *utils.UploadResult, uploadParams UploadParams) artifactContext {
	return func(artifact UploadData) parallel.TaskFunc {
		return func(threadId int) (e error) {
			if artifact.IsDir {
				us.createFolderInArtifactory(artifact)
				return
			}
			var uploaded bool
			var target string
			var artifactFileInfo utils.FileInfo
			uploadResult.TotalCount[threadId]++
			logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, us.DryRun)
			target, e = utils.BuildArtifactoryUrl(us.ArtDetails.GetUrl(), artifact.Artifact.TargetPath, make(map[string]string))
			if e != nil {
				return
			}
			artifactFileInfo, uploaded, e = us.uploadFile(artifact.Artifact.LocalPath, target, artifact.Artifact.TargetPath, artifact.Props, uploadParams, logMsgPrefix)
			if e != nil {
				return
			}
			if uploaded {
				uploadResult.SuccessCount[threadId]++
				uploadResult.FileInfo[threadId] = append(uploadResult.FileInfo[threadId], artifactFileInfo)
			}
			return
		}
	}
}

func (us *UploadService) createFolderInArtifactory(artifact UploadData) error {
	url, err := utils.BuildArtifactoryUrl(us.ArtDetails.GetUrl(), artifact.Artifact.TargetPath, make(map[string]string))
	url = clientutils.AddTrailingSlashIfNeeded(url)
	if err != nil {
		return err
	}
	content := make([]byte, 0)
	httpClientsDetails := us.ArtDetails.CreateHttpClientDetails()
	resp, body, err := us.client.SendPut(url, content, &httpClientsDetails)
	if err != nil {
		log.Debug(resp)
		return err
	}
	logUploadResponse("Uploaded directory:", resp, body, false, us.DryRun)
	return err
}

func NewUploadParams() UploadParams {
	return UploadParams{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{}, MinChecksumDeploy: 10240}
}

func getVcsProps(path string, vcsCache *clientutils.VcsCache) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	props := ""
	revision, url, err := vcsCache.GetVcsDetails(filepath.Dir(path))
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	if revision != "" {
		props += ";vcs.revision=" + revision
	}
	if url != "" {
		props += ";vcs.url=" + url
	}
	return props, nil
}
