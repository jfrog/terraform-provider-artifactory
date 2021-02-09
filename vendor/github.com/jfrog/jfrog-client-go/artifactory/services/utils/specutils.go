package utils

import (
	"strings"
)

const (
	WILDCARD SpecType = "wildcard"
	AQL      SpecType = "aql"
	BUILD    SpecType = "build"
)

type SpecType string

type Aql struct {
	ItemsFind string `json:"items.find"`
}

type ArtifactoryCommonParams struct {
	Aql     Aql
	Pattern string
	// Deprecated, use Exclusions instead
	ExcludePatterns []string
	Exclusions      []string
	Target          string
	Props           string
	ExcludeProps    string
	SortOrder       string
	SortBy          []string
	Offset          int
	Limit           int
	Build           string
	Bundle          string
	Recursive       bool
	IncludeDirs     bool
	Regexp          bool
	ArchiveEntries  string
}

type FileGetter interface {
	GetAql() Aql
	GetPattern() string
	SetPattern(pattern string)
	GetExclusions() []string
	// Deprecated, Use Exclusions instead
	GetExcludePatterns() []string
	GetTarget() string
	SetTarget(target string)
	IsExplode() bool
	GetProps() string
	GetSortOrder() string
	GetSortBy() []string
	GetOffset() int
	GetLimit() int
	GetBuild() string
	GetBundle() string
	GetSpecType() (specType SpecType)
	IsRegexp() bool
	IsRecursive() bool
	IsIncludeDirs() bool
	GetArchiveEntries() string
	SetArchiveEntries(archiveEntries string)
}

func (params ArtifactoryCommonParams) GetArchiveEntries() string {
	return params.ArchiveEntries
}

func (params *ArtifactoryCommonParams) SetArchiveEntries(archiveEntries string) {
	params.ArchiveEntries = archiveEntries
}

func (params *ArtifactoryCommonParams) GetPattern() string {
	return params.Pattern
}

func (params *ArtifactoryCommonParams) SetPattern(pattern string) {
	params.Pattern = pattern
}

func (params *ArtifactoryCommonParams) SetTarget(target string) {
	params.Target = target
}

func (params *ArtifactoryCommonParams) GetTarget() string {
	return params.Target
}

func (params *ArtifactoryCommonParams) GetProps() string {
	return params.Props
}

func (params *ArtifactoryCommonParams) GetExcludeProps() string {
	return params.ExcludeProps
}

func (params *ArtifactoryCommonParams) IsExplode() bool {
	return params.Recursive
}

func (params *ArtifactoryCommonParams) IsRecursive() bool {
	return params.Recursive
}

func (params *ArtifactoryCommonParams) IsRegexp() bool {
	return params.Regexp
}

func (params *ArtifactoryCommonParams) GetAql() Aql {
	return params.Aql
}

func (params *ArtifactoryCommonParams) GetBuild() string {
	return params.Build
}

func (params *ArtifactoryCommonParams) GetBundle() string {
	return params.Bundle
}

func (params ArtifactoryCommonParams) IsIncludeDirs() bool {
	return params.IncludeDirs
}

func (params *ArtifactoryCommonParams) SetProps(props string) {
	params.Props = props
}

func (params *ArtifactoryCommonParams) SetExcludeProps(excludeProps string) {
	params.ExcludeProps = excludeProps
}

func (params *ArtifactoryCommonParams) GetSortBy() []string {
	return params.SortBy
}

func (params *ArtifactoryCommonParams) GetSortOrder() string {
	return params.SortOrder
}

func (params *ArtifactoryCommonParams) GetOffset() int {
	return params.Offset
}

func (params *ArtifactoryCommonParams) GetLimit() int {
	return params.Limit
}

func (params *ArtifactoryCommonParams) GetExcludePatterns() []string {
	return params.ExcludePatterns
}

func (params *ArtifactoryCommonParams) GetExclusions() []string {
	return params.Exclusions
}

func (aql *Aql) UnmarshalJSON(value []byte) error {
	str := string(value)
	first := strings.Index(str[strings.Index(str, "{")+1:], "{")
	last := strings.LastIndex(str, "}")

	aql.ItemsFind = str[first+1 : last]
	return nil
}

func (params ArtifactoryCommonParams) GetSpecType() (specType SpecType) {
	switch {
	case params.Build != "" && params.Aql.ItemsFind == "" && (params.Pattern == "*" || params.Pattern == ""):
		specType = BUILD
	case params.Aql.ItemsFind != "":
		specType = AQL
	default:
		specType = WILDCARD
	}
	return specType
}
