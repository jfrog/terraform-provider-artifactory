package utils

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type (
	VcsCache struct {
		// Key - Path to the .git directory.
		// Value - Reference to a struct, storing the URL and revision.
		vcsRootDir sync.Map
		// Key - Path to a directory.
		// Value - Reference to a struct, storing the URL and revision from the upstream .git. Can also include nil, if there's no upstream .git.
		vcsDir sync.Map
		// The current size of vcsDir
		vcsDirSize *int32 // Size of vcs folders entries
	}
	vcsDetails struct {
		url      string
		revision string
	}
)

const MAX_ENTRIES = 10000

func NewVcsDetals() *VcsCache {
	return &VcsCache{vcsRootDir: sync.Map{}, vcsDir: sync.Map{}, vcsDirSize: new(int32)}
}

func (this *VcsCache) incCacheSize(num int32) {
	atomic.AddInt32(this.vcsDirSize, num)
}

func (this *VcsCache) getCacheSize() int32 {
	return atomic.LoadInt32(this.vcsDirSize)
}

// Search for '.git' directory inside 'path', incase there is one, extract the details and add a new entry to the cache(key:path in the file system ,value: git revision & url).
// otherwise, search in the parent folder and try:
// 1. search for .git, and save the details for the current dir and all subpath
// 2. .git not found, go to parent dir and repeat
// 3. not found on the root directory, add all subpath to cache with nil as a value
func (this *VcsCache) GetVcsDetails(path string) (revision, refUrl string, err error) {
	keys := strings.Split(path, string(os.PathSeparator))
	var subPath string
	var subPaths []string
	var vcsDetailsResult *vcsDetails
	for i := len(keys); i > 0; i-- {
		subPath = strings.Join(keys[:i], string(os.PathSeparator))
		// Try to get from cache
		if vcsDetails, found := this.searchCache(subPath); found {
			if vcsDetails != nil {
				revision, refUrl, vcsDetailsResult = vcsDetails.revision, vcsDetails.url, vcsDetails
			}
			break
		}
		// Begin directory search
		revision, refUrl, err = tryGetGitDetails(subPath, this)
		if revision != "" || refUrl != "" {
			vcsDetailsResult = &vcsDetails{revision: revision, url: refUrl}
			this.vcsRootDir.Store(subPath, vcsDetailsResult)
			break
		}
		if err != nil {
			return
		}
		subPaths = append(subPaths, subPath)
	}
	if size := len(subPaths); size > 0 {
		this.clearCacheIfExceedsMax()
		for _, v := range subPaths {
			this.vcsDir.Store(v, vcsDetailsResult)
		}
		this.incCacheSize(int32(size))
	}
	return
}

func (this *VcsCache) clearCacheIfExceedsMax() {
	if this.getCacheSize() > MAX_ENTRIES {
		this.vcsDir = sync.Map{}
		this.vcsDirSize = new(int32)
	}
}

func tryGetGitDetails(path string, this *VcsCache) (string, string, error) {
	exists, err := fileutils.IsDirExists(filepath.Join(path, ".git"), false)
	if exists {
		return extractGitDetails(path)
	}
	return "", "", err
}

func extractGitDetails(path string) (string, string, error) {
	gitService := NewGitManager(path)
	if err := gitService.ReadConfig(); err != nil {
		return "", "", err
	}
	return gitService.GetRevision(), gitService.GetUrl(), nil
}

func (this *VcsCache) searchCache(path string) (*vcsDetails, bool) {
	if data, found := this.vcsDir.Load(path); found {
		if vcsDetails, ok := data.(*vcsDetails); ok {
			return vcsDetails, ok
		}
	}
	if data, found := this.vcsRootDir.Load(path); found {
		if vcsDetails, ok := data.(*vcsDetails); ok {
			return vcsDetails, ok
		}
	}
	return nil, false
}
