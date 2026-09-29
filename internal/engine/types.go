package engine

import "time"

type Category string

const (
	CategoryAll       Category = "All files"
	CategoryImages    Category = "Images"
	CategoryVideos    Category = "Videos"
	CategoryAudio     Category = "Audio"
	CategoryDocuments Category = "Documents"
	CategoryArchives  Category = "Archives"
	CategoryCustom    Category = "Custom"
)

type Method string

const (
	MethodTemplate     Method = "template"
	MethodReplace      Method = "replace"
	MethodPrefixSuffix Method = "prefix_suffix"
	MethodCase         Method = "case"
)

type CaseMode string

const (
	CaseLower CaseMode = "lower"
	CaseUpper CaseMode = "upper"
	CaseTitle CaseMode = "title"
)

type Config struct {
	Root             string
	Recursive        bool
	Category         Category
	CustomExtensions string
	Method           Method
	Template         string
	Find             string
	ReplaceWith      string
	UseRegex         bool
	Prefix           string
	Suffix           string
	CaseMode         CaseMode
	BatchTime        time.Time
}

type Item struct {
	SourcePath  string
	Folder      string
	OldName     string
	NewName     string
	Status      string
	Error       string
	Checked     bool
	DirIndex    int
	GlobalIndex int
}

const (
	StatusOK        = "OK"
	StatusUnchanged = "Unchanged"
	StatusConflict  = "Conflict"
	StatusInvalid   = "Invalid"
)
