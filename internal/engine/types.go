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

type CollisionPolicy string

const (
	CollisionSkip       CollisionPolicy = "skip"
	CollisionAutoNumber CollisionPolicy = "auto-number"
	CollisionOverwrite  CollisionPolicy = "overwrite"
	CollisionStop       CollisionPolicy = "stop"
)

type SortMode string

const (
	SortName      SortMode = "name"
	SortCreated   SortMode = "created"
	SortModified  SortMode = "modified"
	SortSize      SortMode = "size"
	SortExtension SortMode = "extension"
	SortPath      SortMode = "path"
	SortAdded     SortMode = "added"
	SortManual    SortMode = "manual"
)

type Method string

const (
	MethodTemplate       Method = "template"
	MethodList           Method = "list"
	MethodListReplace    Method = "list_replace"
	MethodReplace        Method = "replace"
	MethodPrefixSuffix   Method = "prefix_suffix"
	MethodCase           Method = "case"
	MethodRemove         Method = "remove"
	MethodRemovePattern  Method = "remove_pattern"
	MethodRenumber       Method = "renumber"
	MethodTrim           Method = "trim"
	MethodTimestamp      Method = "timestamp"
	MethodMove           Method = "move"
	MethodSwap           Method = "swap"
	MethodScript         Method = "script"
)

type CaseMode string

const (
	CaseLower CaseMode = "lower"
	CaseUpper CaseMode = "upper"
	CaseTitle CaseMode = "title"
)

const (
	PositionPrefix = "prefix"
	PositionSuffix = "suffix"

	TimestampModified = "modified"
	TimestampBatch    = "batch"
)

// RenameMethod describes one step in a rename pipeline.
// Methods are applied from top to bottom to the name produced by the previous step.
type RenameMethod struct {
	Type     Method `json:"type"`
	Disabled bool   `json:"disabled,omitempty"`

	Template string `json:"template,omitempty"`

	ListText             string `json:"list_text,omitempty"`
	ListIncludeExtension bool   `json:"list_include_extension,omitempty"`

	ListReplaceText          string `json:"list_replace_text,omitempty"`
	ListReplaceRegex         bool   `json:"list_replace_regex,omitempty"`
	ListReplaceCaseSensitive bool   `json:"list_replace_case_sensitive,omitempty"`

	Find        string `json:"find,omitempty"`
	ReplaceWith string `json:"replace_with,omitempty"`
	UseRegex    bool   `json:"use_regex,omitempty"`

	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`

	CaseMode CaseMode `json:"case_mode,omitempty"`

	RemoveStart int `json:"remove_start,omitempty"`
	RemoveCount int `json:"remove_count,omitempty"`

	RemovePattern      string `json:"remove_pattern,omitempty"`
	RemovePatternRegex bool   `json:"remove_pattern_regex,omitempty"`

	RenumberStart     int    `json:"renumber_start,omitempty"`
	RenumberStep      int    `json:"renumber_step,omitempty"`
	RenumberPadding   int    `json:"renumber_padding,omitempty"`
	RenumberPerDir    bool   `json:"renumber_per_dir,omitempty"`
	RenumberPosition  string `json:"renumber_position,omitempty"`
	RenumberSeparator string `json:"renumber_separator,omitempty"`

	TrimNormalizeSpaces bool `json:"trim_normalize_spaces,omitempty"`

	TimestampSource    string `json:"timestamp_source,omitempty"`
	TimestampFormat    string `json:"timestamp_format,omitempty"`
	TimestampPosition  string `json:"timestamp_position,omitempty"`
	TimestampSeparator string `json:"timestamp_separator,omitempty"`

	MoveStart int `json:"move_start,omitempty"`
	MoveCount int `json:"move_count,omitempty"`
	MoveTo    int `json:"move_to,omitempty"`

	SwapSeparator  string `json:"swap_separator,omitempty"`
	SwapOccurrence int    `json:"swap_occurrence,omitempty"`

	ScriptExpression string `json:"script_expression,omitempty"`
}

type Config struct {
	// Sources may contain any mix of files and directories.
	// Root is kept for compatibility with older callers and is used when Sources is empty.
	Sources       []string
	Root          string
	ExcludedPaths []string

	Recursive        bool
	Category         Category
	CustomExtensions string

	SortBy              SortMode
	SortDescending      bool
	SortPerFolder       bool
	CollisionPolicy     CollisionPolicy
	SkipImageDimensions bool

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

	Size     int64
	Width    int
	Height   int
	Created  time.Time
	Modified time.Time
}

const (
	StatusOK        = "OK"
	StatusUnchanged = "Unchanged"
	StatusConflict  = "Conflict"
	StatusInvalid   = "Invalid"
)
