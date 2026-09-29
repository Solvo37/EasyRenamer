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

// RenameMethod describes one step in a rename pipeline.
// Methods are applied from top to bottom to the name produced by the previous step.
type RenameMethod struct {
	Type        Method
	Template    string
	Find        string
	ReplaceWith string
	UseRegex    bool
	Prefix      string
	Suffix      string
	CaseMode    CaseMode
}

type Config struct {
	// Sources may contain any mix of files and directories.
	// Root is kept for compatibility with older callers and is used when Sources is empty.
	Sources []string
	Root    string

	Recursive        bool
	Category         Category
	CustomExtensions string

	// Methods is the preferred API. If it is empty, the legacy single-method
	// fields below are used so older integrations keep working.
	Methods []RenameMethod

	Method      Method
	Template    string
	Find        string
	ReplaceWith string
	UseRegex    bool
	Prefix      string
	Suffix      string
	CaseMode    CaseMode
	BatchTime   time.Time
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
