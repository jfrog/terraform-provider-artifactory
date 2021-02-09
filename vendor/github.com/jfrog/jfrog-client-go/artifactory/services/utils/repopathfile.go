package utils

import (
	"github.com/jfrog/jfrog-client-go/utils"
	"regexp"
	"strings"
)

// We need to translate the provided download pattern to an AQL query.
// In Artifactory, for each artifact the name and path of the artifact are saved separately.
// We therefore need to build an AQL query that covers all possible repositories, paths and names the provided
// pattern can include.
// For example, the pattern repo/a/* can include the two following file:
// repo/a/file1.tgz and also repo/a/b/file2.tgz
// To achieve that, this function parses the pattern by splitting it by its * characters.
// The end result is a list of RepoPathFile structs.
// Each struct represent a possible repository, path and file name triple to be included in AQL query with an "or" relationship.
type RepoPathFile struct {
	repo string
	path string
	file string
}

var asteriskRegexp = regexp.MustCompile(`\*`)

func createRepoPathFileTriples(pattern string, recursive bool) []RepoPathFile {
	firstSlashIndex := strings.Index(pattern, "/")
	asteriskIndices := asteriskRegexp.FindAllStringIndex(pattern, -1)

	if asteriskIndices != nil && !utils.IsSlashPrecedeAsterisk(asteriskIndices[0][0], firstSlashIndex) {
		var triples []RepoPathFile
		var lastRepoAsteriskIndex int
		for _, asteriskIndex := range asteriskIndices {
			if utils.IsSlashPrecedeAsterisk(asteriskIndex[0], firstSlashIndex) {
				break
			}
			repo := pattern[:asteriskIndex[0]+1]     // '<repo>*'
			newPattern := pattern[asteriskIndex[0]:] // '*<pattern>'
			newPattern = strings.TrimPrefix(newPattern, "*/")
			triples = append(triples, createPathFilePairs(repo, newPattern, recursive)...)
			lastRepoAsteriskIndex = asteriskIndex[1]
		}

		// Handle characters between last asterisk before first slash: "a*handle-it/"
		if lastRepoAsteriskIndex < firstSlashIndex {
			repo := pattern[:firstSlashIndex]         // '<repo>*'
			newPattern := pattern[firstSlashIndex+1:] // '*<pattern>'
			triples = append(triples, createPathFilePairs(repo, newPattern, recursive)...)
		} else if firstSlashIndex < 0 && !strings.HasSuffix(pattern, "*") {
			// Handle characters after last asterisk "a*handle-it"
			triples = append(triples, createPathFilePairs(pattern, "*", recursive)...)
		}
		return triples
	}

	if firstSlashIndex < 0 {
		return createPathFilePairs(pattern, "*", recursive)
	}
	repo := pattern[:firstSlashIndex]
	pattern = pattern[firstSlashIndex+1:]
	return createPathFilePairs(repo, pattern, recursive)
}

func createPathFilePairs(repo, pattern string, recursive bool) []RepoPathFile {
	if pattern == "*" {
		return []RepoPathFile{{repo, getDefaultPath(recursive), "*"}}
	}

	path, name, triples := handleNonRecursiveTriples(repo, pattern, recursive)
	if !recursive {
		return triples
	}
	if name == "*" {
		return append(triples, RepoPathFile{repo, path + "/*", "*"})
	}

	nameSplit := strings.Split(name, "*")
	for i := 0; i < len(nameSplit)-1; i++ {
		str := ""
		for j, namePart := range nameSplit {
			if j > 0 {
				str += "*"
			}
			if j == i {
				str += nameSplit[i] + "*/"
			} else {
				str += namePart
			}
		}
		slashSplit := strings.Split(str, "/")
		filePath := slashSplit[0]
		fileName := slashSplit[1]
		if fileName == "" {
			fileName = "*"
		}
		if path != "" && !strings.HasSuffix(path, "/") {
			path += "/"
		}
		triples = append(triples, RepoPathFile{repo, path + filePath, fileName})
	}
	return triples
}

func handleNonRecursiveTriples(repo, pattern string, recursive bool) (string, string, []RepoPathFile) {
	slashIndex := strings.LastIndex(pattern, "/")
	if slashIndex < 0 {
		// Optimization - If pattern starts with `*`, we'll have a triple with <repo>*<file>.
		// In that case we'd prefer to avoid <repo>.<file>.
		if recursive && strings.HasPrefix(pattern, "*") {
			return "", pattern, []RepoPathFile{}
		}
		return "", pattern, []RepoPathFile{{repo, ".", pattern}}
	}
	path := pattern[:slashIndex]
	name := pattern[slashIndex+1:]
	return path, name, []RepoPathFile{{repo, path, name}}
}

func getDefaultPath(recursive bool) string {
	if recursive {
		return "*"
	}
	return "."
}
