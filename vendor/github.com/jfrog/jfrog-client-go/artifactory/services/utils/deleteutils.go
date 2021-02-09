package utils

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
)

func WildcardToDirsPath(deletePattern, searchResult string) (string, error) {
	if !strings.HasSuffix(deletePattern, "/") {
		return "", errors.New("Delete pattern must end with \"/\"")
	}

	regexpPattern := "^" + strings.Replace(deletePattern, "*", "([^/]*|.*)", -1)
	r, err := regexp.Compile(regexpPattern)
	errorutils.CheckError(err)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(searchResult)
	if len(groups) > 0 {
		return groups[0], nil
	}
	return "", nil
}

// Write all the dirs to be deleted into 'resultWriter'.
// However, skip dirs with files(s) that should not be deleted.
// In order to accomplish this, we check if the dirs are a prefix of any artifact, witch means the folder contains the artifact and should not be deleted.
// Optimization: In order not to scan for each dir the entire artifact reader and see if it is a prefix or not, we rely on the fact that the dirs and artifacts are sorted.
// We have two sorted readers in ascending order, we will start scanning from the beginning of the lists and compare whether the folder is a prefix of the current artifact,
// in case this is true the dir should not be deleted and we can move on to the next dir, otherwise we have to continue to the next dir or artifact.
// To know this, we will choose to move on with the lexicographic largest between the two.
//
// candidateDirsReaders - Sorted list of dirs to be deleted.
// filesNotToBeDeleteReader - Sorted files that should not be deleted.
// resultWriter - The filtered list of dirs to be deleted.
func WriteCandidateDirsToBeDeleted(candidateDirsReaders []*content.ContentReader, filesNotToBeDeleteReader *content.ContentReader, resultWriter *content.ContentWriter) (err error) {
	dirsToBeDeletedReader, err := MergeSortedFiles(candidateDirsReaders, true)
	if err != nil {
		return
	}
	defer dirsToBeDeletedReader.Close()
	var candidateDirToBeDeletedPath string
	var artifactNotToBeDeletePath string
	var candidateDirToBeDeleted, artifactNotToBeDeleted *ResultItem
	for {
		// Fetch the next 'candidateDirToBeDeleted'.
		if candidateDirToBeDeleted == nil {
			candidateDirToBeDeleted = new(ResultItem)
			if err = dirsToBeDeletedReader.NextRecord(candidateDirToBeDeleted); err != nil {
				break
			}
			if candidateDirToBeDeleted.Name == "." {
				continue
			}
			candidateDirToBeDeletedPath = strings.ToLower(candidateDirToBeDeleted.GetItemRelativePath())
		}
		// Fetch the next 'artifactNotToBeDelete'.
		if artifactNotToBeDeleted == nil {
			artifactNotToBeDeleted = new(ResultItem)
			if err = filesNotToBeDeleteReader.NextRecord(artifactNotToBeDeleted); err != nil {
				// No artifacts left, write remaining dirs to be deleted to result file.
				resultWriter.Write(*candidateDirToBeDeleted)
				writeRemainCandidate(resultWriter, dirsToBeDeletedReader)
				break
			}
			artifactNotToBeDeletePath = strings.ToLower(artifactNotToBeDeleted.GetItemRelativePath())
		}
		// Found an 'artifact not to be deleted' in 'dir to be deleted', therefore skip writing the dir to the result file.
		if strings.HasPrefix(artifactNotToBeDeletePath, candidateDirToBeDeletedPath) {
			candidateDirToBeDeleted = nil
			continue
		}
		// 'artifactNotToBeDeletePath' & 'candidateDirToBeDeletedPath' are both sorted. As a result 'candidateDirToBeDeleted' cant be a prefix for any of the remaining artifacts.
		if artifactNotToBeDeletePath > candidateDirToBeDeletedPath {
			resultWriter.Write(*candidateDirToBeDeleted)
			candidateDirToBeDeleted = nil
			continue
		}
		artifactNotToBeDeleted = nil
	}
	err = filesNotToBeDeleteReader.GetError()
	filesNotToBeDeleteReader.Reset()
	return
}

func writeRemainCandidate(cw *content.ContentWriter, mergeResult *content.ContentReader) {
	for toBeDeleted := new(ResultItem); mergeResult.NextRecord(toBeDeleted) == nil; toBeDeleted = new(ResultItem) {
		cw.Write(*toBeDeleted)
	}
}

func FilterCandidateToBeDeleted(deleteCandidates *content.ContentReader, resultWriter *content.ContentWriter) ([]*content.ContentReader, error) {
	paths := make(map[string]ResultItem)
	pathsKeys := make([]string, 0, utils.MaxBufferSize)
	dirsToBeDeleted := []*content.ContentReader{}
	for candidate := new(ResultItem); deleteCandidates.NextRecord(candidate) == nil; candidate = new(ResultItem) {
		// Save all dirs candidate in a diffrent temp file.
		if candidate.Type == "folder" {
			if candidate.Name == "." {
				continue
			}
			pathsKeys = append(pathsKeys, candidate.GetItemRelativePath())
			paths[candidate.GetItemRelativePath()] = *candidate
			if len(pathsKeys) == utils.MaxBufferSize {
				sortedCandidateDirsFile, err := SortAndSaveBufferToFile(paths, pathsKeys, true)
				if err != nil {
					return nil, err
				}
				dirsToBeDeleted = append(dirsToBeDeleted, sortedCandidateDirsFile)
				// Init buffer.
				paths = make(map[string]ResultItem)
				pathsKeys = make([]string, 0, utils.MaxBufferSize)
			}
		} else {
			// Write none dir results.
			resultWriter.Write(*candidate)
		}
	}
	if err := deleteCandidates.GetError(); err != nil {
		return nil, err
	}
	deleteCandidates.Reset()
	if len(pathsKeys) > 0 {
		sortedFile, err := SortAndSaveBufferToFile(paths, pathsKeys, true)
		if err != nil {
			return nil, err
		}
		dirsToBeDeleted = append(dirsToBeDeleted, sortedFile)
	}
	return dirsToBeDeleted, nil
}
