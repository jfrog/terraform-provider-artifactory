package utils

import (
	"bufio"
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"path/filepath"
	"strings"
)

type manager struct {
	path     string
	err      error
	revision string
	url      string
}

func NewGitManager(path string) *manager {
	dotGitPath := filepath.Join(path, ".git")
	return &manager{path: dotGitPath}
}

func (m *manager) ReadConfig() error {
	if m.path == "" {
		return errorutils.CheckError(errors.New(".git path must be defined."))
	}
	m.readRevision()
	m.readUrl()
	return m.err
}

func (m *manager) GetUrl() string {
	return m.url
}

func (m *manager) GetRevision() string {
	return m.revision
}

func (m *manager) readUrl() {
	if m.err != nil {
		return
	}
	dotGitPath := filepath.Join(m.path, "config")
	file, err := os.Open(dotGitPath)
	if errorutils.CheckError(err) != nil {
		m.err = err
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var IsNextLineUrl bool
	var originUrl string
	for scanner.Scan() {
		if IsNextLineUrl {
			text := scanner.Text()
			strings.HasPrefix(text, "url")
			originUrl = strings.TrimSpace(strings.SplitAfter(text, "=")[1])
			break
		}
		if scanner.Text() == "[remote \"origin\"]" {
			IsNextLineUrl = true
		}
	}
	if err := scanner.Err(); err != nil {
		errorutils.CheckError(err)
		m.err = err
		return
	}
	if !strings.HasSuffix(originUrl, ".git") {
		originUrl += ".git"
	}
	m.url = originUrl

	// Mask url if required
	regExp, err := GetRegExp(CredentialsInUrlRegexp)
	if err != nil {
		m.err = err
		return
	}
	matchedResult := regExp.FindString(originUrl)
	if matchedResult == "" {
		return
	}
	m.url = MaskCredentials(originUrl, matchedResult)
}

func (m *manager) getRevisionOrBranchPath() (revision, refUrl string, err error) {
	dotGitPath := filepath.Join(m.path, "HEAD")
	file, e := os.Open(dotGitPath)
	if errorutils.CheckError(e) != nil {
		err = e
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "ref") {
			refUrl = strings.TrimSpace(strings.SplitAfter(text, ":")[1])
			break
		}
		revision = text
	}
	if err = scanner.Err(); err != nil {
		errorutils.CheckError(err)
	}
	return
}

func (m *manager) readRevision() {
	if m.err != nil {
		return
	}
	// This function will either return the revision or the branch ref:
	revision, ref, err := m.getRevisionOrBranchPath()
	if err != nil {
		m.err = err
		return
	}
	// If the revision was returned, then we're done:
	if revision != "" {
		m.revision = revision
		return
	}

	// Else, if found ref try getting revision using it.
	refPath := filepath.Join(m.path, ref)
	exists, err := fileutils.IsFileExists(refPath, false)
	if err != nil {
		m.err = err
		return
	}
	if exists {
		m.readRevisionFromRef(refPath)
		return
	}
	// Otherwise, try to find .git/packed-refs and look for the HEAD there
	m.readRevisionFromPackedRef(ref)
}

func (m *manager) readRevisionFromRef(refPath string) {
	revision := ""
	file, err := os.Open(refPath)
	if errorutils.CheckError(err) != nil {
		m.err = err
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		revision = strings.TrimSpace(text)
		break
	}
	if err := scanner.Err(); err != nil {
		m.err = errorutils.CheckError(err)
		return
	}

	m.revision = revision
	return
}

func (m *manager) readRevisionFromPackedRef(ref string) {
	packedRefPath := filepath.Join(m.path, "packed-refs")
	exists, err := fileutils.IsFileExists(packedRefPath, false)
	if err != nil {
		m.err = err
		return
	}
	if exists {
		file, err := os.Open(packedRefPath)
		if errorutils.CheckError(err) != nil {
			m.err = err
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Expecting to find the revision (the full extended SHA-1, or a unique leading substring) followed by the ref.
			if strings.HasSuffix(line, ref) {
				split := strings.Split(line, " ")
				if len(split) == 2 {
					m.revision = split[0]
				} else {
					m.err = errorutils.CheckError(errors.New("failed fetching revision for ref :" + ref + " - Unexpected line structure in packed-refs file"))
				}
				return
			}
		}
		if err = scanner.Err(); err != nil {
			m.err = errorutils.CheckError(err)
			return
		}
	}

	m.err = errorutils.CheckError(errors.New("failed fetching revision from git config, from ref: " + ref))
	return
}
