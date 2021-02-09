package utils

import (
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils"
)

// Returns an AQL body string to search file in Artifactory by pattern, according the the specified arguments requirements.
func CreateAqlBodyForSpecWithPattern(params *ArtifactoryCommonParams) (string, error) {
	searchPattern := prepareSourceSearchPattern(params.Pattern, params.Target, true)
	repoPathFileTriples := createRepoPathFileTriples(searchPattern, params.Recursive)
	includeRoot := strings.Count(searchPattern, "/") < 2
	triplesSize := len(repoPathFileTriples)

	propsQueryPart, err := buildPropsQueryPart(params.Props, params.ExcludeProps)
	if err != nil {
		return "", err
	}
	itemTypeQuery := buildItemTypeQueryPart(params)
	nePath := buildNePathPart(triplesSize == 0 || includeRoot)
	excludeQuery := buildExcludeQueryPart(params, triplesSize == 0 || params.Recursive, params.Recursive)
	releaseBundle, err := buildReleaseBundleQuery(params)
	if err != nil {
		return "", err
	}

	json := fmt.Sprintf(`{%s"$or":[`, propsQueryPart+itemTypeQuery+nePath+excludeQuery+releaseBundle)

	// Get archive search parameters
	archivePathFilePairs := createArchiveSearchParams(params)

	json += handleRepoPathFileTriples(repoPathFileTriples, archivePathFilePairs, triplesSize) + "]}"
	return json, nil
}

func createArchiveSearchParams(params *ArtifactoryCommonParams) []RepoPathFile {
	var archivePathFilePairs []RepoPathFile

	if params.ArchiveEntries != "" {
		archiveSearchPattern := prepareSearchPattern(params.ArchiveEntries, false)
		archivePathFilePairs = createPathFilePairs("", archiveSearchPattern, true)
	}

	return archivePathFilePairs
}

// Handle building aql query when having PathFilePairs
func handleRepoPathFileTriples(pathFilePairs []RepoPathFile, archivePathFilePairs []RepoPathFile, pathPairSize int) string {
	var query string
	archivePathPairSize := len(archivePathFilePairs)

	for i := 0; i < pathPairSize; i++ {
		if archivePathPairSize > 0 {
			query += handleArchiveSearch(pathFilePairs[i], archivePathFilePairs)
		} else {
			query += buildInnerQueryPart(pathFilePairs[i])
		}

		if i+1 < pathPairSize {
			query += ","
		}
	}

	return query
}

// Handle building aql query including archive search
func handleArchiveSearch(triple RepoPathFile, archivePathFilePairs []RepoPathFile) string {
	var query string
	archivePathPairSize := len(archivePathFilePairs)
	for i := 0; i < archivePathPairSize; i++ {
		query += buildInnerArchiveQueryPart(triple, archivePathFilePairs[i].path, archivePathFilePairs[i].file)

		if i+1 < archivePathPairSize {
			query += ","
		}
	}
	return query
}

func createAqlBodyForBuild(buildName, buildNumber string) string {
	itemsPart :=
		`{` +
			`"artifact.module.build.name":"%s",` +
			`"artifact.module.build.number":"%s"` +
			`}`
	return fmt.Sprintf(itemsPart, buildName, buildNumber)
}

func createAqlQueryForBuild(buildName, buildNumber, includeQueryPart string) string {
	queryBody := createAqlBodyForBuild(buildName, buildNumber)
	itemsPart := `items.find(%s)%s`
	return fmt.Sprintf(itemsPart, queryBody, includeQueryPart)
}

//noinspection GoUnusedExportedFunction
func CreateAqlQueryForNpm(npmName, npmVersion string) string {
	itemsPart :=
		`items.find({` +
			`"@npm.name":"%s",` +
			`"@npm.version":"%s"` +
			`})%s`
	return fmt.Sprintf(itemsPart, npmName, npmVersion, buildIncludeQueryPart([]string{"name", "repo", "path", "actual_sha1", "actual_md5"}))
}

func CreateAqlQueryForPypi(repo, file string) string {
	itemsPart :=
		`items.find({` +
			`"repo": "%s",` +
			`"$or": [{` +
			`"$and":[{` +
			`"path": {"$match": "*"},` +
			`"name": {"$match": "%s"}` +
			`}]` +
			`}]` +
			`})%s`
	return fmt.Sprintf(itemsPart, repo, file, buildIncludeQueryPart([]string{"name", "repo", "path", "actual_md5", "actual_sha1"}))
}

func prepareSearchPattern(pattern string, repositoryExists bool) string {
	addWildcardIfNeeded(&pattern, repositoryExists)
	// Remove parenthesis
	pattern = strings.Replace(pattern, "(", "", -1)
	pattern = strings.Replace(pattern, ")", "", -1)
	return pattern
}

func buildPropsQueryPart(props, excludeProps string) (string, error) {
	propsQuery := ""
	properties, err := ParseProperties(props, JoinCommas)
	if err != nil {
		return "", err
	}
	for _, v := range properties.Properties {
		propsQuery += buildKeyValQueryPart(v.Key, v.Value) + `,`
	}

	excludePropsQuery := ""
	excludeProperties, err := ParseProperties(excludeProps, JoinCommas)
	if err != nil {
		return "", err
	}
	excludePropsLen := len(excludeProperties.Properties)
	if excludePropsLen == 1 {
		singleProp := &excludeProperties.Properties[0]
		excludePropsQuery = buildExcludedKeyValQueryPart(singleProp.Key, singleProp.Value) + `,`
	} else if excludePropsLen > 1 {
		excludePropsQuery = `"$or":[`
		for _, v := range excludeProperties.Properties {
			excludePropsQuery += `{` + buildExcludedKeyValQueryPart(v.Key, v.Value) + `},`
		}
		excludePropsQuery = strings.TrimSuffix(excludePropsQuery, ",") + `],`
	}
	return propsQuery + excludePropsQuery, nil
}

func buildKeyValQueryPart(key string, value string) string {
	return fmt.Sprintf(`"@%s":%s`, key, getAqlValue(value))
}

func buildExcludedKeyValQueryPart(key string, value string) string {
	return fmt.Sprintf(`"@%s":{"$ne":%s}`, key, getAqlValue(value))
}

func buildItemTypeQueryPart(params *ArtifactoryCommonParams) string {
	if params.IncludeDirs {
		return `"type":"any",`
	}
	return ""
}

func buildNePathPart(includeRoot bool) string {
	if !includeRoot {
		return `"path":{"$ne":"."},`
	}
	return ""
}

func buildInnerQueryPart(triple RepoPathFile) string {
	innerQueryPattern := `{"$and":` +
		`[{` +
		`"repo":%s,` +
		`"path":%s,` +
		`"name":%s` +
		`}]}`
	return fmt.Sprintf(innerQueryPattern, getAqlValue(triple.repo), getAqlValue(triple.path), getAqlValue(triple.file))
}

func buildInnerArchiveQueryPart(triple RepoPathFile, archivePath, archiveName string) string {
	innerQueryPattern := `{"$and":` +
		`[{` +
		`"repo":%s,` +
		`"path":%s,` +
		`"name":%s,` +
		`"archive.entry.path":%s,` +
		`"archive.entry.name":%s` +
		`}]}`
	return fmt.Sprintf(innerQueryPattern, getAqlValue(triple.repo), getAqlValue(triple.path), getAqlValue(triple.file), getAqlValue(archivePath), getAqlValue(archiveName))
}

func buildExcludeQueryPart(params *ArtifactoryCommonParams, useLocalPath, recursive bool) string {
	excludeQuery := ""
	var excludeTriples []RepoPathFile
	if len(params.GetExclusions()) > 0 {
		for _, excludePattern := range params.GetExclusions() {
			excludeTriples = append(excludeTriples, createRepoPathFileTriples(prepareSearchPattern(excludePattern, true), recursive)...)
		}
	} else {
		// Support legacy exclude patterns. 'Exclude patterns' are deprecated and replaced by 'exclusions'.
		for _, excludePattern := range params.GetExcludePatterns() {
			excludeTriples = append(excludeTriples, createPathFilePairs("", prepareSearchPattern(excludePattern, false), recursive)...)
		}
	}

	for _, excludeTriple := range excludeTriples {
		excludePath := excludeTriple.path
		if !useLocalPath && excludePath == "." {
			excludePath = "*"
		}
		excludeRepoStr := ""
		if excludeTriple.repo != "" {
			excludeRepoStr = fmt.Sprintf(`"repo":{"$nmatch":"%s"},`, excludeTriple.repo)
		}
		excludeQuery += fmt.Sprintf(`"$or":[{%s"path":{"$nmatch":"%s"},"name":{"$nmatch":"%s"}}],`, excludeRepoStr, excludePath, excludeTriple.file)
	}
	return excludeQuery
}

func buildReleaseBundleQuery(params *ArtifactoryCommonParams) (string, error) {
	bundleName, bundleVersion, err := parseNameAndVersion(params.Bundle, false)
	if bundleName == "" || err != nil {
		return "", err
	}
	itemsPart := `"$and":` +
		`[{` +
		`"release_artifact.release.name":%s,` +
		`"release_artifact.release.version":%s` +
		`}],`
	return fmt.Sprintf(itemsPart, getAqlValue(bundleName), getAqlValue(bundleVersion)), nil
}

// Creates a list of basic required return fields. The list will include the sortBy field if needed.
// If requiredArtifactProps is NONE or 'includePropertiesInAqlForSpec' return false,
// "property" field won't be included due to a limitation in the AQL implementation in Artifactory.
func getQueryReturnFields(specFile *ArtifactoryCommonParams, requiredArtifactProps RequiredArtifactProps) []string {
	returnFields := []string{"name", "repo", "path", "actual_md5", "actual_sha1", "size", "type", "modified", "created"}
	if !includePropertiesInAqlForSpec(specFile) {
		// Sort dose not work when property is in the include section. in this case we will append properties in later stage.
		return appendMissingFields(specFile.SortBy, returnFields)
	}
	if requiredArtifactProps != NONE {
		// If any prop is needed we just add all the properties to the result.
		return append(returnFields, "property")
	}
	return returnFields
}

// If specFile includes sortBy or limit, the produced AQL won't include property in the include section.
// This due to an Artifactory limitation related to using these flags with props in an AQL statement.
// Meaning - the result won't contain properties.
func includePropertiesInAqlForSpec(specFile *ArtifactoryCommonParams) bool {
	return !(len(specFile.SortBy) > 0 || specFile.Limit > 0)
}

func appendMissingFields(fields []string, defaultFields []string) []string {
	for _, field := range fields {
		if !fileutils.IsStringInSlice(field, defaultFields) {
			defaultFields = append(defaultFields, field)
		}
	}
	return defaultFields
}

func prepareFieldsForQuery(fields []string) []string {
	// Since a slice is basically a pointer, we don't want to modify the underlying fields array because it might be used again (like in delete service)
	// We will create new slice with the quoted values and will return it.
	var queryFields []string
	for _, val := range fields {
		queryFields = append(queryFields, `"`+val+`"`)
	}
	return queryFields
}

// Creates an aql query from a spec file.
func BuildQueryFromSpecFile(specFile *ArtifactoryCommonParams, requiredArtifactProps RequiredArtifactProps) string {
	aqlBody := specFile.Aql.ItemsFind
	query := fmt.Sprintf(`items.find(%s)%s`, aqlBody, buildIncludeQueryPart(getQueryReturnFields(specFile, requiredArtifactProps)))
	query = appendSortQueryPart(specFile, query)
	query = appendOffsetQueryPart(specFile, query)
	return appendLimitQueryPart(specFile, query)
}

func appendOffsetQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if specFile.Offset > 0 {
		query = fmt.Sprintf(`%s.offset(%s)`, query, strconv.Itoa(specFile.Offset))
	}
	return query
}

func appendLimitQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if specFile.Limit > 0 {
		query = fmt.Sprintf(`%s.limit(%s)`, query, strconv.Itoa(specFile.Limit))
	}
	return query
}

func appendSortQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if len(specFile.SortBy) > 0 {
		query = fmt.Sprintf(`%s.sort({%s})`, query, buildSortQueryPart(specFile.SortBy, specFile.SortOrder))
	}
	return query
}

func buildSortQueryPart(sortFields []string, sortOrder string) string {
	if sortOrder == "" {
		sortOrder = "asc"
	}
	return fmt.Sprintf(`"$%s":[%s]`, sortOrder, strings.Join(prepareFieldsForQuery(sortFields), `,`))
}

func createPropsQuery(aqlBody, propKey, propVal string) string {
	propKeyValQueryPart := buildKeyValQueryPart(propKey, propVal)
	propsQuery :=
		`items.find({` +
			`"$and":[%s,{%s}]` +
			`})%s`
	return fmt.Sprintf(propsQuery, aqlBody, propKeyValQueryPart, buildIncludeQueryPart([]string{"name", "repo", "path", "actual_sha1", "property"}))
}

func buildIncludeQueryPart(fieldsToInclude []string) string {
	fieldsToInclude = prepareFieldsForQuery(fieldsToInclude)
	return fmt.Sprintf(`.include(%s)`, strings.Join(fieldsToInclude, `,`))
}

// Optimization - If value is a wildcard pattern, return `{"$match":"value"}`. Otherwise, return `"value"`.
func getAqlValue(val string) string {
	var aqlValuePattern string
	if strings.Contains(val, "*") {
		aqlValuePattern = `{"$match":"%s"}`
	} else {
		aqlValuePattern = `"%s"`
	}
	return fmt.Sprintf(aqlValuePattern, val)
}

func prepareSourceSearchPattern(pattern, target string, repositoryExists bool) string {
	addWildcardIfNeeded(&pattern, repositoryExists)
	pattern = utils.RemovePlaceholderParentheses(pattern, target)
	return pattern
}

func addWildcardIfNeeded(pattern *string, repositoryExists bool) {
	if strings.HasSuffix(*pattern, "/") || (*pattern == "" && repositoryExists) {
		*pattern += "*"
	}
}
